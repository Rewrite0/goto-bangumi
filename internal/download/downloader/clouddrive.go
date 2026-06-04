package downloader

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	clouddrive "goto-bangumi/gen"
	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/model"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// CloudDriveDownloader implements BaseDownloader using CloudDrive2 gRPC API.
// It maps torrent offline downloads onto the standard downloader interface:
//   - "hash"     = torrent info hash (same meaning as qBittorrent)
//   - "savePath" = CloudDrive2 virtual path, e.g. /115Open/Downloads
//
// CloudDrive2 endpoint is the gRPC address, default port 19798.
type CloudDriveDownloader struct {
	conn        *grpc.ClientConn
	rpc         clouddrive.CloudDriveFileSrvClient
	config      *model.DownloaderConfig
	mu          sync.RWMutex
	token       string
	cloudName   string
	cloudAcctID string
	apiInterval int
}

func NewCloudDriveDownloader() *CloudDriveDownloader {
	return &CloudDriveDownloader{apiInterval: 500}
}

// Init creates the gRPC connection. Call Auth afterwards.
func (d *CloudDriveDownloader) Init(config *model.DownloaderConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	d.config = config

	var cred grpc.DialOption
	if config.Ssl {
		cred = grpc.WithTransportCredentials(credentials.NewTLS(nil))
	} else {
		cred = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.NewClient(config.Host, cred)
	if err != nil {
		return fmt.Errorf("CloudDrive2 connect to %s: %w", config.Host, err)
	}
	d.conn = conn
	d.rpc = clouddrive.NewCloudDriveFileSrvClient(conn)
	return nil
}

// authCtx attaches the stored JWT token to outgoing gRPC metadata.
func (d *CloudDriveDownloader) authCtx(ctx context.Context) context.Context {
	d.mu.RLock()
	tok := d.token
	d.mu.RUnlock()
	if tok == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+tok)
}

// Auth authenticates against CloudDrive2 and stores the JWT token.
func (d *CloudDriveDownloader) Auth(ctx context.Context) (bool, error) {
	resp, err := d.rpc.GetToken(ctx, &clouddrive.GetTokenRequest{
		UserName: d.config.Username,
		Password: d.config.Password,
	})
	if err != nil {
		slog.Error("[CloudDrive2] auth error", "host", d.config.Host, "error", err)
		return false, &apperrors.NetworkError{Err: fmt.Errorf("CloudDrive2 auth: %w", err)}
	}
	if !resp.GetSuccess() {
		slog.Error("[CloudDrive2] auth failed", "message", resp.GetErrorMessage())
		return false, &apperrors.DownloadAuthenticationError{
			Err:  fmt.Errorf("CloudDrive2 login failed: %s", resp.GetErrorMessage()),
			Name: d.config.Username,
		}
	}

	d.mu.Lock()
	d.token = resp.GetToken()
	d.mu.Unlock()

	slog.Info("[CloudDrive2] auth success", "user", d.config.Username)
	d.resolveCloudInfo(ctx)
	return true, nil
}

// resolveCloudInfo fetches cloud API list and matches cloudName/cloudAcctID
// from the first path segment of SavePath (e.g. /115Open/Downloads → "115Open").
func (d *CloudDriveDownloader) resolveCloudInfo(ctx context.Context) {
	aCtx := d.authCtx(ctx)
	apis, err := d.rpc.GetAllCloudApis(aCtx, &emptypb.Empty{})
	if err != nil || len(apis.GetApis()) == 0 {
		slog.Warn("[CloudDrive2] could not fetch cloud APIs", "error", err)
		return
	}

	targetName := cloudNameFromPath(d.config.SavePath)
	for _, api := range apis.GetApis() {
		if api.GetName() == targetName {
			d.mu.Lock()
			d.cloudName = api.GetName()
			d.cloudAcctID = api.GetUserName()
			d.mu.Unlock()
			slog.Debug("[CloudDrive2] resolved cloud", "name", d.cloudName, "accountId", d.cloudAcctID)
			return
		}
	}

	// Fallback: use the first available cloud
	first := apis.GetApis()[0]
	d.mu.Lock()
	d.cloudName = first.GetName()
	d.cloudAcctID = first.GetUserName()
	d.mu.Unlock()
	slog.Debug("[CloudDrive2] fallback cloud", "name", d.cloudName, "accountId", d.cloudAcctID)
}

// cloudNameFromPath extracts the first path segment: "/115Open/dir" → "115Open".
func cloudNameFromPath(path string) string {
	p := strings.TrimPrefix(path, "/")
	if idx := strings.IndexByte(p, '/'); idx > 0 {
		return p[:idx]
	}
	return p
}

// Logout logs out from CloudDrive2 and closes the connection.
func (d *CloudDriveDownloader) Logout(ctx context.Context) (bool, error) {
	_, err := d.rpc.Logout(d.authCtx(ctx), &clouddrive.UserLogoutRequest{
		LogoutFromCloudFS: false,
	})
	d.mu.Lock()
	d.token = ""
	d.mu.Unlock()
	if d.conn != nil {
		_ = d.conn.Close()
		d.conn = nil
	}
	if err != nil {
		return false, fmt.Errorf("CloudDrive2 logout: %w", err)
	}
	return true, nil
}

// Add submits a magnet URI to CloudDrive2 offline download queue.
// Returns the info hashes already known from TorrentInfo since CloudDrive2
// does not echo them back in the response.
func (d *CloudDriveDownloader) Add(ctx context.Context, torrentInfo *model.TorrentInfo, savePath string) ([]string, error) {
	url := torrentInfo.MagnetURI
	if url == "" {
		return nil, fmt.Errorf("[CloudDrive2] magnet URI required for offline download")
	}

	result, err := d.rpc.AddOfflineFiles(d.authCtx(ctx), &clouddrive.AddOfflineFileRequest{
		Urls:     url,
		ToFolder: savePath,
	})
	if err != nil {
		return nil, &apperrors.NetworkError{Err: fmt.Errorf("CloudDrive2 AddOfflineFiles: %w", err)}
	}
	if !result.GetSuccess() {
		return nil, fmt.Errorf("[CloudDrive2] add offline failed: %s", result.GetErrorMessage())
	}

	var hashes []string
	if torrentInfo.InfoHashV1 != "" {
		hashes = append(hashes, torrentInfo.InfoHashV1)
	}
	if torrentInfo.InfoHashV2 != "" {
		v2 := torrentInfo.InfoHashV2
		if len(v2) > 40 {
			v2 = v2[:40]
		}
		hashes = append(hashes, v2)
	}
	return hashes, nil
}

// listOfflineFiles returns all offline download tasks under SavePath.
func (d *CloudDriveDownloader) listOfflineFiles(ctx context.Context) ([]*clouddrive.OfflineFile, error) {
	result, err := d.rpc.ListOfflineFilesByPath(d.authCtx(ctx), &clouddrive.FileRequest{
		Path: d.config.SavePath,
	})
	if err != nil {
		return nil, fmt.Errorf("CloudDrive2 ListOfflineFilesByPath: %w", err)
	}
	return result.GetOfflineFiles(), nil
}

// GetTorrentInfo looks up an offline download task by its info hash.
func (d *CloudDriveDownloader) GetTorrentInfo(ctx context.Context, hash string) (*model.TorrentDownloadInfo, error) {
	files, err := d.listOfflineFiles(ctx)
	if err != nil {
		return nil, err
	}

	hashLower := strings.ToLower(hash)
	for _, f := range files {
		if strings.ToLower(f.GetInfoHash()) == hashLower {
			info := &model.TorrentDownloadInfo{
				SavePath: d.config.SavePath,
				ETA:      -1,
			}
			if f.GetStatus() == clouddrive.OfflineFileStatus_OFFLINE_FINISHED {
				info.ETA = 0
				info.Completed = int(time.Now().Unix())
			}
			return info, nil
		}
	}
	return nil, &apperrors.DownloadKeyError{
		Err: fmt.Errorf("offline task not found"),
		Key: hash,
	}
}

// GetTorrentFiles lists video/subtitle files in the save folder corresponding
// to the named offline download identified by hash.
func (d *CloudDriveDownloader) GetTorrentFiles(ctx context.Context, hash string) ([]string, error) {
	offlineFiles, err := d.listOfflineFiles(ctx)
	if err != nil {
		return nil, err
	}

	hashLower := strings.ToLower(hash)
	var taskName string
	for _, f := range offlineFiles {
		if strings.ToLower(f.GetInfoHash()) == hashLower {
			taskName = f.GetName()
			break
		}
	}
	if taskName == "" {
		return nil, &apperrors.DownloadKeyError{
			Err: fmt.Errorf("offline task not found"),
			Key: hash,
		}
	}

	folderPath := d.config.SavePath + "/" + taskName
	subFiles, err := d.listSubFiles(ctx, folderPath)
	if err != nil {
		// Task may still be downloading; return no files rather than error
		slog.Warn("[CloudDrive2] GetTorrentFiles: could not list folder", "path", folderPath, "error", err)
		return nil, nil
	}

	names := make([]string, 0, len(subFiles))
	for _, f := range subFiles {
		if !f.GetIsDirectory() {
			names = append(names, taskName+"/"+f.GetName())
		}
	}
	return names, nil
}

// listSubFiles collects all file entries from the streaming GetSubFiles RPC.
func (d *CloudDriveDownloader) listSubFiles(ctx context.Context, path string) ([]*clouddrive.CloudDriveFile, error) {
	stream, err := d.rpc.GetSubFiles(d.authCtx(ctx), &clouddrive.ListSubFileRequest{
		Path:         path,
		ForceRefresh: false,
	})
	if err != nil {
		return nil, err
	}

	var all []*clouddrive.CloudDriveFile
	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		all = append(all, reply.GetSubFiles()...)
	}
	return all, nil
}

// TorrentsInfo returns all offline download tasks as generic maps.
// statusFilter, category, tag, limit are accepted for interface compatibility
// but CloudDrive2 does not support server-side filtering on offline tasks.
func (d *CloudDriveDownloader) TorrentsInfo(ctx context.Context, statusFilter, category string, tag *string, limit int) ([]map[string]any, error) {
	files, err := d.listOfflineFiles(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, 0, len(files))
	for i, f := range files {
		if limit > 0 && i >= limit {
			break
		}
		entry := map[string]any{
			"hash":         f.GetInfoHash(),
			"name":         f.GetName(),
			"size":         f.GetSize(),
			"status":       f.GetStatus().String(),
			"progress":     f.GetPercendDone(),
			"save_path":    d.config.SavePath,
			"category":     "GotoBangumi",
		}
		result = append(result, entry)
	}
	return result, nil
}

// CheckHash verifies an info hash exists in the offline download list.
func (d *CloudDriveDownloader) CheckHash(ctx context.Context, hash string) (string, error) {
	files, err := d.listOfflineFiles(ctx)
	if err != nil {
		return hash, err
	}
	hashLower := strings.ToLower(hash)
	for _, f := range files {
		if strings.ToLower(f.GetInfoHash()) == hashLower {
			return hash, nil
		}
	}
	return hash, &apperrors.DownloadKeyError{Err: fmt.Errorf("hash not found"), Key: hash}
}

// Rename renames a file at oldPath to newPath on the CloudDrive2 filesystem.
// torrentHash is unused; oldPath and newPath must be full CloudDrive2 virtual paths.
func (d *CloudDriveDownloader) Rename(ctx context.Context, torrentHash, oldPath, newPath string) (bool, error) {
	// newPath should be just the new filename, not the full path
	newName := newPath
	if idx := strings.LastIndexByte(newPath, '/'); idx >= 0 {
		newName = newPath[idx+1:]
	}

	result, err := d.rpc.RenameFile(d.authCtx(ctx), &clouddrive.RenameFileRequest{
		TheFilePath: oldPath,
		NewName:     newName,
	})
	if err != nil {
		return false, &apperrors.NetworkError{Err: fmt.Errorf("CloudDrive2 RenameFile: %w", err)}
	}
	if !result.GetSuccess() {
		return false, fmt.Errorf("[CloudDrive2] rename failed: %s", result.GetErrorMessage())
	}
	return true, nil
}

// Move moves files (identified by their CloudDrive2 paths) to newLocation.
func (d *CloudDriveDownloader) Move(ctx context.Context, hashes []string, newLocation string) (bool, error) {
	result, err := d.rpc.MoveFile(d.authCtx(ctx), &clouddrive.MoveFileRequest{
		TheFilePaths: hashes,
		DestPath:     newLocation,
	})
	if err != nil {
		return false, &apperrors.NetworkError{Err: fmt.Errorf("CloudDrive2 MoveFile: %w", err)}
	}
	if !result.GetSuccess() {
		return false, fmt.Errorf("[CloudDrive2] move failed: %s", result.GetErrorMessage())
	}
	return true, nil
}

// Delete removes offline download tasks by their info hashes.
func (d *CloudDriveDownloader) Delete(ctx context.Context, hashes []string) (bool, error) {
	d.mu.RLock()
	cloudName := d.cloudName
	cloudAcctID := d.cloudAcctID
	d.mu.RUnlock()

	result, err := d.rpc.RemoveOfflineFiles(d.authCtx(ctx), &clouddrive.RemoveOfflineFilesRequest{
		CloudName:      cloudName,
		CloudAccountId: cloudAcctID,
		InfoHashes:     hashes,
		DeleteFiles:    true,
	})
	if err != nil {
		return false, &apperrors.NetworkError{Err: fmt.Errorf("CloudDrive2 RemoveOfflineFiles: %w", err)}
	}
	if !result.GetSuccess() {
		return false, fmt.Errorf("[CloudDrive2] delete failed: %s", result.GetErrorMessage())
	}
	return true, nil
}

// GetDownloadURL returns the HTTP download URL for a file at path.
// The caller is responsible for making the HTTP request with the returned headers.
func (d *CloudDriveDownloader) GetDownloadURL(ctx context.Context, filePath string, preview bool) (string, map[string]string, error) {
	info, err := d.rpc.GetDownloadUrlPath(d.authCtx(ctx), &clouddrive.GetDownloadUrlPathRequest{
		Path:         filePath,
		Preview:      preview,
		GetDirectUrl: true,
	})
	if err != nil {
		return "", nil, fmt.Errorf("CloudDrive2 GetDownloadUrlPath: %w", err)
	}

	// Prefer direct URL from cloud storage provider
	if directURL := info.GetDirectUrl(); directURL != "" {
		headers := info.GetAdditionalHeaders()
		if ua := info.GetUserAgent(); ua != "" {
			if headers == nil {
				headers = make(map[string]string)
			}
			headers["User-Agent"] = ua
		}
		return directURL, headers, nil
	}

	// Assemble URL from template: replace {SCHEME}, {HOST}, {PREVIEW}
	scheme := "http"
	if d.config.Ssl {
		scheme = "https"
	}
	host := d.config.Host
	previewStr := "false"
	if preview {
		previewStr = "true"
	}

	urlPath := info.GetDownloadUrlPath()
	urlPath = strings.ReplaceAll(urlPath, "{SCHEME}", scheme)
	urlPath = strings.ReplaceAll(urlPath, "{HOST}", host)
	urlPath = strings.ReplaceAll(urlPath, "{PREVIEW}", previewStr)

	assembledURL := scheme + "://" + host + urlPath
	return assembledURL, nil, nil
}

func (d *CloudDriveDownloader) GetInterval() int {
	return d.apiInterval
}
