package apperrors

import "errors"

// DownloadKeyError 不能找到对应的 key
type DownloadKeyError struct {
	Err error
	Key string
}

func (e *DownloadKeyError) Error() string {
	return "download key error: " + e.Err.Error()
}

func (e *DownloadKeyError) Unwrap() error {
	return e.Err
}

func IsKeyError(err error) bool {
	var keyErr *DownloadKeyError
	return errors.As(err, &keyErr)
}

// DownloadAuthenticationError 目前没有认证
type DownloadAuthenticationError struct {
	Err  error
	Name string
}

func (e *DownloadAuthenticationError) Error() string {
	return "download authentication error: " + e.Err.Error()
}

func (e *DownloadAuthenticationError) Unwrap() error {
	return e.Err
}

func IsDownloadAuthenticationError(err error) bool {
	var authErr *DownloadAuthenticationError
	return errors.As(err, &authErr)
}

// DownloadForbiddenError 操作下载器禁止
type DownloadForbiddenError struct {
	Err error
}

func (e *DownloadForbiddenError) Error() string {
	return "download forbidden error: " + e.Err.Error()
}

func (e *DownloadForbiddenError) Unwrap() error {
	return e.Err
}

func IsDownloadForbiddenError(err error) bool {
	var forbiddenErr *DownloadForbiddenError
	return errors.As(err, &forbiddenErr)
}

// DownloadLoginError 表明当前登录有问题, 不应在短时间内重试, 需要用户检查配置,手动处理
type DownloadLoginError struct {
	Err error
}

func (e *DownloadLoginError) Error() string {
	return "download login error: " + e.Err.Error()
}

func (e *DownloadLoginError) Unwrap() error {
	return e.Err
}

func NewDownloadLoginError(err error) *DownloadLoginError {
	return &DownloadLoginError{Err: err}
}

func IsDownloadLoginError(err error) bool {
	var loginErr *DownloadLoginError
	return errors.As(err, &loginErr)
}
