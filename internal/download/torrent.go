package download

import (
	"bytes"
	"encoding/hex"

	"goto-bangumi/internal/model"

	"github.com/anacrolix/torrent/metainfo"
)

func ParseTorrent(torrent []byte) (*model.TorrentInfo, error) {
	// mikan_url := "https://mikanani.me/Download/20250515/32587df888ce2b3f7d9df67854ea10e50153a55c.torrent"
	//
	mi, err := metainfo.Load(bytes.NewReader(torrent))
	if err != nil {
		return nil, err
	}
	magnetV2, err := mi.MagnetV2()
	if err != nil {
		return nil, err
	}
	return parse(&magnetV2, torrent)
}

func ParseTorrentURL(torrentURL string) (*model.TorrentInfo, error) {
	magnetV2, err := metainfo.ParseMagnetV2Uri(torrentURL)
	if err != nil {
		return nil, err
	}
	return parse(&magnetV2, nil)
}

func parse(magnetV2 *metainfo.MagnetV2, torrent []byte) (*model.TorrentInfo, error) {
	v1Hash := ""
	if v1, ok := magnetV2.InfoHash.AsTuple(); ok {
		v1Hash = v1.HexString()
	}
	v2Hash := ""
	if v2, ok := magnetV2.V2InfoHash.AsTuple(); ok {
		v2Hash = hex.EncodeToString(v2[:])
	}
	ti := &model.TorrentInfo{
		Name:       magnetV2.DisplayName,
		InfoHashV1: v1Hash,
		InfoHashV2: v2Hash,
		MagnetURI:  magnetV2.String(),
		File:       torrent,
	}
	return ti, nil
}
