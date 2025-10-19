package apperrors

import "errors"


// NetworkError 网络错误类型
type NetworkError struct {
	Err        error
	StatusCode int // HTTP 状态码，0 表示非 HTTP 错误
}

func (e *NetworkError) Error() string {
	if e.StatusCode > 0 {
		return "network error: " + e.Err.Error()
	}
	return "network error: " + e.Err.Error()
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// ParseError 解析错误类型
type ParseError struct {
	Err error
}

func (e *ParseError) Error() string {
	return "parse error: " + e.Err.Error()
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// IsNetworkError 判断是否为网络错误
func IsNetworkError(err error) bool {
	var netErr *NetworkError
	return errors.As(err, &netErr)
}

// IsParseError 判断是否为解析错误
func IsParseError(err error) bool {
	var parseErr *ParseError
	return errors.As(err, &parseErr)
}

// IsUnauthorizedError 判断是否为 401 未授权错误
func IsUnauthorizedError(err error) bool {
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return netErr.StatusCode == 401
	}
	return false
}

// GetStatusCode 从 NetworkError 中获取 HTTP 状态码，如果不是 NetworkError 则返回 0
func GetStatusCode(err error) int {
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return netErr.StatusCode
	}
	return 0
}
