package apperrors

import "errors"

// 定义解析器相关错误类型
var (
	// ErrNetwork 网络请求相关错误
	ErrNetwork = errors.New("network error")

	// ErrParse 解析内容相关错误
	ErrParse = errors.New("parse error")
)

// NetworkError 网络错误类型
type NetworkError struct {
	Err error
}

func (e *NetworkError) Error() string {
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
