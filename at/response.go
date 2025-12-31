package at

import (
	"strings"
)

// ResponseSet 定义可配置的命令最终响应类型集合
type ResponseSet struct {
	// 基本响应
	OK    string // 成功响应
	Error string // 错误响应

	// 通话相关
	NoCarrier  string // 无载波
	NoAnswer   string // 无应答
	NoDialtone string // 无拨号音
	Busy       string // 忙线
	Connect    string // 连接成功

	// 错误相关
	CMEError string // CME 错误（设备错误）
	CMSError string // CMS 错误（短信服务错误）

	// 提示符
	Prompt string // 输入提示符（如 SMS 的 ">"）

	// 自定义响应
	CustomFinal []string // 自定义最终响应列表
}

// DefaultResponseSet 返回默认的命令响应类型集合
func DefaultResponseSet() *ResponseSet {
	return &ResponseSet{
		// 基本响应
		OK:    "OK",
		Error: "ERROR",

		// 通话相关
		NoCarrier:  "NO CARRIER",
		NoAnswer:   "NO ANSWER",
		NoDialtone: "NO DIALTONE",
		Busy:       "BUSY",
		Connect:    "CONNECT",

		// 错误相关
		CMEError: "+CME ERROR:",
		CMSError: "+CMS ERROR:",

		// 提示符
		Prompt: ">",

		// 自定义响应
		CustomFinal: []string{},
	}
}

// GetAllResponses 返回所有最终响应的列表
func (rs *ResponseSet) GetAllResponses() []string {
	responses := []string{
		// 基本响应
		rs.OK,
		rs.Error,

		// 通话相关
		rs.NoCarrier,
		rs.NoAnswer,
		rs.NoDialtone,
		rs.Busy,
		rs.Connect,

		// 错误相关
		rs.CMEError,
		rs.CMSError,

		// 提示符
		rs.Prompt,
	}
	// 添加自定义最终响应
	return append(responses, rs.CustomFinal...)
}

// IsFinal 检查是否为最终响应
func (rs *ResponseSet) IsFinal(line string) bool {
	for _, resp := range rs.GetAllResponses() {
		if resp != "" && strings.HasPrefix(line, resp) {
			return true
		}
	}
	return false
}

// IsError 检查是否为错误响应
func (rs *ResponseSet) IsError(line string) bool {
	responses := []string{
		rs.Error,
		rs.NoCarrier,
		rs.NoAnswer,
		rs.NoDialtone,
		rs.Busy,
		rs.CMEError,
		rs.CMSError,
	}
	for _, resp := range responses {
		if resp != "" && strings.HasPrefix(line, resp) {
			return true
		}
	}
	return false
}

// IsSuccess 检查是否为成功响应
func (rs *ResponseSet) IsSuccess(line string) bool {
	responses := []string{
		rs.OK,
		rs.Connect,
		rs.Prompt,
	}
	for _, resp := range responses {
		if resp != "" && strings.HasPrefix(line, resp) {
			return true
		}
	}
	return false
}
