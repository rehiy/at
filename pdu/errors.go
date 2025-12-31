package pdu

import "fmt"

// Error 定义 PDU 错误类型
type Error struct {
	Code    ErrorCode
	Message string
	Err     error
}

// ErrorCode 错误代码
type ErrorCode int

const (
	// ErrorCodeInvalidPDU 无效的 PDU 格式
	ErrorCodeInvalidPDU ErrorCode = iota + 1
	// ErrorCodeInvalidEncoding 无效的编码
	ErrorCodeInvalidEncoding
	// ErrorCodeInvalidPhoneNumber 无效的电话号码
	ErrorCodeInvalidPhoneNumber
	// ErrorCodeInvalidSMSC 无效的短信中心号码
	ErrorCodeInvalidSMSC
	// ErrorCodeMessageTooLong 消息过长
	ErrorCodeMessageTooLong
	// ErrorCodeUnsupportedFeature 不支持的特性
	ErrorCodeUnsupportedFeature
	// ErrorCodeInvalidUDH 无效的用户数据头
	ErrorCodeInvalidUDH
	// ErrorCodeInvalidTimestamp 无效的时间戳
	ErrorCodeInvalidTimestamp
)

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("PDU error [%d]: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("PDU error [%d]: %s", e.Code, e.Message)
}

// Unwrap 返回底层错误
func (e *Error) Unwrap() error {
	return e.Err
}

// NewError 创建新的 PDU 错误
func NewError(code ErrorCode, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
