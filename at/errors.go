package at

import (
	"errors"
	"fmt"
)

// 错误定义
var (
	// 连接相关错误
	ErrConnectionClosed  = errors.New("connection is closed")
	ErrPortNotAvailable  = errors.New("serial port not available")
	ErrInvalidPortConfig = errors.New("invalid port configuration")

	// 命令相关错误
	ErrCommandTimeout     = errors.New("command timeout")
	ErrInvalidCommand     = errors.New("invalid AT command")
	ErrUnexpectedResponse = errors.New("unexpected response from modem")

	// 响应相关错误
	ErrNoResponse    = errors.New("no response from modem")
	ErrResponseParse = errors.New("failed to parse response")
	ErrResponseError = errors.New("modem returned error response")

	// 通知相关错误
	ErrNotificationFailed = errors.New("notification listening failed")
	ErrHandlerNotSet      = errors.New("notification handler not set")
)

// ATError 表示 AT 命令执行错误
type ATError struct {
	Command    string
	Response   []string
	Underlying error
}

// Error 返回错误信息
func (e *ATError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("AT command %q failed: %v (response: %v)", e.Command, e.Underlying, e.Response)
	}
	return fmt.Sprintf("AT command %q failed with response: %v", e.Command, e.Response)
}

// Unwrap 返回底层错误
func (e *ATError) Unwrap() error {
	return e.Underlying
}

// NewATError 创建一个新的 ATError
func NewATError(command string, response []string, underlying error) *ATError {
	return &ATError{
		Command:    command,
		Response:   response,
		Underlying: underlying,
	}
}

// IsConnectionError 判断是否为连接错误
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionClosed) ||
		errors.Is(err, ErrPortNotAvailable) ||
		errors.Is(err, ErrInvalidPortConfig)
}

// IsCommandError 判断是否为命令错误
func IsCommandError(err error) bool {
	return errors.Is(err, ErrCommandTimeout) ||
		errors.Is(err, ErrInvalidCommand) ||
		errors.Is(err, ErrUnexpectedResponse)
}

// IsResponseError 判断是否为响应错误
func IsResponseError(err error) bool {
	return errors.Is(err, ErrNoResponse) ||
		errors.Is(err, ErrResponseParse) ||
		errors.Is(err, ErrResponseError)
}
