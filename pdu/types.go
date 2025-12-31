package pdu

import "time"

// MessageType 定义消息类型
type MessageType int

const (
	// MessageTypeSMSDeliver 接收的短信
	MessageTypeSMSDeliver MessageType = 0x00
	// MessageTypeSMSSubmit 发送的短信
	MessageTypeSMSSubmit MessageType = 0x01
	// MessageTypeSMSStatusReport 状态报告
	MessageTypeSMSStatusReport MessageType = 0x02
)

// Encoding 定义编码类型
type Encoding int

const (
	// Encoding7Bit GSM 7-bit 默认字母表
	Encoding7Bit Encoding = 0x00
	// Encoding8Bit 8-bit 数据
	Encoding8Bit Encoding = 0x04
	// EncodingUCS2 UCS2（Unicode）编码
	EncodingUCS2 Encoding = 0x08
)

// ValidityPeriod 定义有效期
type ValidityPeriod byte

const (
	// ValidityPeriod1Hour 1 小时
	ValidityPeriod1Hour ValidityPeriod = 0x0B
	// ValidityPeriod6Hours 6 小时
	ValidityPeriod6Hours ValidityPeriod = 0x47
	// ValidityPeriod24Hours 24 小时
	ValidityPeriod24Hours ValidityPeriod = 0xA7
	// ValidityPeriod1Week 1 周
	ValidityPeriod1Week ValidityPeriod = 0xAD
	// ValidityPeriodMaximum 最大有效期（63 周）
	ValidityPeriodMaximum ValidityPeriod = 0xFF
)

// Message 表示一条短信
type Message struct {
	// Type 消息类型
	Type MessageType
	// PhoneNumber 电话号码
	PhoneNumber string
	// Text 短信内容
	Text string
	// Encoding 编码类型
	Encoding Encoding
	// SMSC 短信中心号码
	SMSC string
	// Timestamp 时间戳（接收消息时使用）
	Timestamp time.Time
	// ValidityPeriod 有效期（发送消息时使用）
	ValidityPeriod ValidityPeriod
	// RequestReport 是否请求状态报告
	RequestReport bool
	// Flash 是否为闪信
	Flash bool
	// UDH 用户数据头
	UDH []byte
	// Reference 消息引用（用于长短信）
	Reference byte
	// Parts 长短信的总部分数
	Parts byte
	// Part 当前部分序号
	Part byte
}

// PDU 表示一个 PDU 字符串及其长度
type PDU struct {
	// Data PDU 数据（十六进制字符串）
	Data string
	// Length TPDU 长度（用于 AT 命令）
	Length int
}

// AddressType 地址类型
type AddressType byte

const (
	// AddressTypeUnknown 未知类型
	AddressTypeUnknown AddressType = 0x81
	// AddressTypeInternational 国际号码
	AddressTypeInternational AddressType = 0x91
	// AddressTypeAlphanumeric 字母数字
	AddressTypeAlphanumeric AddressType = 0xD0
)

// 常量定义
const (
	// MaxSingleSMSLength 单条短信最大长度（7-bit 编码）
	MaxSingleSMSLength = 160
	// MaxSingleSMSLengthUCS2 单条短信最大长度（UCS2 编码）
	MaxSingleSMSLengthUCS2 = 70
	// MaxConcatSMSLength 长短信单部分最大长度（7-bit 编码）
	MaxConcatSMSLength = 153
	// MaxConcatSMSLengthUCS2 长短信单部分最大长度（UCS2 编码）
	MaxConcatSMSLengthUCS2 = 67
	// UDHConcatSMS8bit 8-bit 引用的长短信 UDH
	UDHConcatSMS8bit = 0x00
	// UDHConcatSMS16bit 16-bit 引用的长短信 UDH
	UDHConcatSMS16bit = 0x08
)
