package at

import (
	"context"
	"strings"
)

// ResponseSet 定义可配置的命令响应类型集合
type ResponseSet struct {
	OK          string   // 成功响应
	Error       string   // 错误响应
	NoCarrier   string   // 无载波
	NoAnswer    string   // 无应答
	NoDialtone  string   // 无拨号音
	Busy        string   // 忙线
	Connect     string   // 连接成功
	CMEError    string   // CME 错误
	CMSError    string   // CMS 错误
	CustomFinal []string // 自定义最终响应列表
}

// DefaultResponseSet 返回默认的命令响应类型集合
func DefaultResponseSet() ResponseSet {
	return ResponseSet{
		OK:          "OK",
		Error:       "ERROR",
		NoCarrier:   "NO CARRIER",
		NoAnswer:    "NO ANSWER",
		NoDialtone:  "NO DIALTONE",
		Busy:        "BUSY",
		Connect:     "CONNECT",
		CMEError:    "+CME ERROR:",
		CMSError:    "+CMS ERROR:",
		CustomFinal: []string{},
	}
}

// GetAllFinalResponses 返回所有最终响应的列表
func (rs *ResponseSet) GetAllFinalResponses() []string {
	responses := []string{
		rs.OK,
		rs.Error,
		rs.NoCarrier,
		rs.NoAnswer,
		rs.NoDialtone,
		rs.Busy,
		rs.Connect,
		rs.CMEError,
		rs.CMSError,
	}

	// 添加自定义最终响应
	return append(responses, rs.CustomFinal...)
}

// IsFinalResponse 检查是否为最终响应
func (rs *ResponseSet) IsFinalResponse(line string) bool {
	for _, resp := range rs.GetAllFinalResponses() {
		if resp != "" && strings.Contains(line, resp) {
			return true
		}
	}
	return false
}

// IsSuccess 检查是否为成功响应
func (rs *ResponseSet) IsSuccess(line string) bool {
	return rs.OK != "" && strings.Contains(line, rs.OK)
}

// IsError 检查是否为错误响应
func (rs *ResponseSet) IsError(line string) bool {
	if rs.Error != "" && strings.Contains(line, rs.Error) {
		return true
	}
	if rs.CMEError != "" && strings.Contains(line, rs.CMEError) {
		return true
	}
	if rs.CMSError != "" && strings.Contains(line, rs.CMSError) {
		return true
	}
	return false
}

// readResponse 从响应通道读取响应
func (m *Connection) readResponse(ctx context.Context) ([]string, error) {
	var responses []string

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case line, ok := <-m.responseChan:
			if !ok {
				// 通道已关闭
				return responses, ErrConnectionClosed
			}

			// 检查是否为最终响应
			if m.responses.IsFinalResponse(line) {
				responses = append(responses, line)
				return responses, nil
			}

			// 添加中间响应
			responses = append(responses, line)
		}
	}
}
