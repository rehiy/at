package at

import (
	"context"
)

// AT 命令通信接口
type AT interface {
	// 发送 AT 命令并等待响应
	SendCommand(ctx context.Context, command string) ([]string, error)

	// 发送 AT 命令并期望特定响应
	SendCommandExpect(ctx context.Context, command string, expected string) error

	// 监听 modem 通知
	ListenNotifications(ctx context.Context, handler NotificationHandler) error

	// 关闭连接
	Close() error

	// 检查连接状态
	IsConnected() bool

	// 测试连接
	Test(ctx context.Context) error

	// 查询制造商信息
	GetManufacturer(ctx context.Context) (string, error)

	// 查询型号信息
	GetModel(ctx context.Context) (string, error)

	// 查询信号质量
	GetSignalQuality(ctx context.Context) (int, int, error)

	// 查询网络状态
	GetNetworkStatus(ctx context.Context) (int, int, error)

	// 设置短信格式为文本模式
	SetSMSFormatText(ctx context.Context) error

	// 设置短信格式为 PDU 模式
	SetSMSFormatPDU(ctx context.Context) error

	// 发送文本短信（自动处理中文和长短信）
	SendSMSText(ctx context.Context, phoneNumber, message string) error

	// 发送 PDU 格式短信
	SendSMSPDU(ctx context.Context, pduData string, length int) error

	// 列出短信
	ListSMS(ctx context.Context, status string) ([]SMS, error)

	// 读取短信
	ReadSMS(ctx context.Context, index int) (*SMS, error)

	// 删除短信
	DeleteSMS(ctx context.Context, index int) error

	// 删除所有短信
	DeleteAllSMS(ctx context.Context) error

	// 拨打电话
	Dial(ctx context.Context, number string) error

	// 挂断电话
	Hangup(ctx context.Context) error

	// 接听电话
	Answer(ctx context.Context) error

	// 重置 modem
	Reset(ctx context.Context) error

	// 关闭回显
	EchoOff(ctx context.Context) error

	// 开启回显
	EchoOn(ctx context.Context) error
}

// NotificationHandler 通知处理函数
type NotificationHandler func(notification string)

// New 创建一个新的 AT 对象
func New(config Config) (AT, error) {
	return newConnection(config)
}
