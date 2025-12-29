package at

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// CommandSet 定义可配置的 AT 命令集
type CommandSet struct {
	// 基本命令
	Test         string // 测试连接
	EchoOff      string // 关闭回显
	EchoOn       string // 开启回显
	Reset        string // 重置 modem
	FactoryReset string // 恢复出厂设置
	SaveSettings string // 保存设置

	// 信息查询
	Manufacturer string // 查询制造商
	Model        string // 查询型号
	Revision     string // 查询版本
	SerialNumber string // 查询序列号
	IMSI         string // 查询 IMSI
	ICCID        string // 查询 ICCID

	// 信号质量
	SignalQuality string // 查询信号质量

	// 网络注册
	NetworkRegistration string // 网络注册状态
	GPRSRegistration    string // GPRS 注册状态

	// 短信相关
	SMSFormat string // 设置短信格式
	ListSMS   string // 列出短信
	ReadSMS   string // 读取短信
	DeleteSMS string // 删除短信
	SendSMS   string // 发送短信

	// 通话相关
	Dial     string // 拨号
	Answer   string // 接听
	Hangup   string // 挂断
	CallerID string // 来电显示
}

// DefaultCommandSet 返回默认的标准 AT 命令集
func DefaultCommandSet() CommandSet {
	return CommandSet{
		// 基本命令
		Test:         "AT",
		EchoOff:      "ATE0",
		EchoOn:       "ATE1",
		Reset:        "ATZ",
		FactoryReset: "AT&F",
		SaveSettings: "AT&W",

		// 信息查询
		Manufacturer: "AT+CGMI",
		Model:        "AT+CGMM",
		Revision:     "AT+CGMR",
		SerialNumber: "AT+CGSN",
		IMSI:         "AT+CIMI",
		ICCID:        "AT+CCID",

		// 信号质量
		SignalQuality: "AT+CSQ",

		// 网络注册
		NetworkRegistration: "AT+CREG",
		GPRSRegistration:    "AT+CGREG",

		// 短信相关
		SMSFormat: "AT+CMGF",
		ListSMS:   "AT+CMGL",
		ReadSMS:   "AT+CMGR",
		DeleteSMS: "AT+CMGD",
		SendSMS:   "AT+CMGS",

		// 通话相关
		Dial:     "ATD",
		Answer:   "ATA",
		Hangup:   "ATH",
		CallerID: "AT+CLIP",
	}
}

// Test 测试连接
func (m *Connection) Test(ctx context.Context) error {
	return m.SendCommandExpect(ctx, m.commands.Test, "OK")
}

// GetManufacturer 查询制造商信息
func (m *Connection) GetManufacturer(ctx context.Context) (string, error) {
	responses, err := m.SendCommand(ctx, m.commands.Manufacturer)
	if err != nil {
		return "", err
	}

	// 查找制造商信息行（不以AT开头的行）
	for _, resp := range responses {
		if !strings.HasPrefix(resp, "AT") {
			return strings.TrimSpace(resp), nil
		}
	}

	return "", fmt.Errorf("no manufacturer info found")
}

// GetModel 查询型号信息
func (m *Connection) GetModel(ctx context.Context) (string, error) {
	responses, err := m.SendCommand(ctx, m.commands.Model)
	if err != nil {
		return "", err
	}

	// 查找型号信息行（不以AT开头的行）
	for _, resp := range responses {
		if !strings.HasPrefix(resp, "AT") {
			return strings.TrimSpace(resp), nil
		}
	}

	return "", fmt.Errorf("no model info found")
}

// GetSignalQuality 查询信号质量
func (m *Connection) GetSignalQuality(ctx context.Context) (int, int, error) {
	responses, err := m.SendCommand(ctx, m.commands.SignalQuality)
	if err != nil {
		return 0, 0, err
	}

	for _, resp := range responses {
		if strings.HasPrefix(resp, "+CSQ:") {
			parts := strings.Split(strings.TrimPrefix(resp, "+CSQ:"), ",")
			if len(parts) >= 2 {
				rssi := parseInt(parts[0])
				ber := parseInt(parts[1])
				return rssi, ber, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("failed to parse signal quality")
}

// GetNetworkStatus 查询网络注册状态
func (m *Connection) GetNetworkStatus(ctx context.Context) (int, int, error) {
	responses, err := m.SendCommand(ctx, m.commands.NetworkRegistration+"?")
	if err != nil {
		return 0, 0, err
	}

	for _, resp := range responses {
		if strings.HasPrefix(resp, "+CREG:") {
			parts := strings.Split(strings.TrimPrefix(resp, "+CREG:"), ",")
			if len(parts) >= 2 {
				n := parseInt(parts[0])
				stat := parseInt(parts[1])
				return n, stat, nil
			}
		}
	}

	return 0, 0, fmt.Errorf("failed to parse network status")
}

// Dial 拨打电话
func (m *Connection) Dial(ctx context.Context, number string) error {
	return m.SendCommandExpect(ctx, m.commands.Dial+number, "OK")
}

// Hangup 挂断电话
func (m *Connection) Hangup(ctx context.Context) error {
	return m.SendCommandExpect(ctx, m.commands.Hangup, "OK")
}

// Answer 接听电话
func (m *Connection) Answer(ctx context.Context) error {
	return m.SendCommandExpect(ctx, m.commands.Answer, "OK")
}

// Reset 重启模块
func (m *Connection) Reset(ctx context.Context) error {
	return m.SendCommandExpect(ctx, m.commands.Reset, "OK")
}

// EchoOff 关闭回显
func (m *Connection) EchoOff(ctx context.Context) error {
	return m.SendCommandExpect(ctx, m.commands.EchoOff, "OK")
}

// EchoOn 开启回显
func (m *Connection) EchoOn(ctx context.Context) error {
	return m.SendCommandExpect(ctx, m.commands.EchoOn, "OK")
}

// parseInt 解析整数
func parseInt(s string) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0 // 保持与原来相同的错误处理行为
	}
	return v
}
