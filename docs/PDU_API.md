# API 文档

本文档详细描述了 rehiy-modem-pdu 库的所有公开 API。

## 目录

- [类型定义](#类型定义)
- [编码函数](#编码函数)
- [解码函数](#解码函数)
- [工具函数](#工具函数)
- [长短信管理](#长短信管理)
- [错误处理](#错误处理)
- [常量定义](#常量定义)

## 类型定义

### Message

`Message` 结构体表示一条短信消息。

```go
type Message struct {
    Type           MessageType     // 消息类型
    PhoneNumber    string          // 电话号码
    Text           string          // 短信内容
    Encoding       Encoding        // 编码类型（0 表示自动选择）
    SMSC           string          // 短信中心号码
    Timestamp      time.Time       // 时间戳（接收消息时使用）
    ValidityPeriod ValidityPeriod  // 有效期（发送消息时使用）
    RequestReport  bool            // 是否请求状态报告
    Flash          bool            // 是否为闪信
    UDH            []byte          // 用户数据头
    Reference      byte            // 消息引用（用于长短信，0 表示自动生成）
    Parts          byte            // 长短信的总部分数
    Part           byte            // 当前部分序号
}
```

**方法**：

```go
func (m *Message) Validate() error
```

验证消息的有效性，检查必填字段和字段值的合法性。

**字段说明**：

- `Type`：消息类型（SMS-DELIVER、SMS-SUBMIT、SMS-STATUS-REPORT）
- `PhoneNumber`：收件人或发件人电话号码
- `Text`：短信文本内容
- `Encoding`：编码方式（自动选择时设为 0）
- `SMSC`：短信中心号码
- `Timestamp`：消息时间戳
- `ValidityPeriod`：消息有效期
- `RequestReport`：是否请求状态报告
- `Flash`：是否为闪信（直接显示不保存）
- `UDH`：用户数据头（长短信使用）
- `Reference`：消息引用号（长短信标识）
- `Parts`：长短信总部分数
- `Part`：当前部分序号

### MessageType

消息类型枚举。

```go
type MessageType int

const (
    MessageTypeSMSDeliver      MessageType = 0x00  // 接收的短信
    MessageTypeSMSSubmit       MessageType = 0x01  // 发送的短信
    MessageTypeSMSStatusReport MessageType = 0x02  // 状态报告
)
```

### Encoding

编码类型枚举。

```go
type Encoding int

const (
    Encoding7Bit Encoding = 0x00  // GSM 7-bit 默认字母表
    Encoding8Bit Encoding = 0x04  // 8-bit 数据
    EncodingUCS2 Encoding = 0x08  // UCS2（Unicode）编码
)
```

### ValidityPeriod

有效期枚举。

```go
type ValidityPeriod byte

const (
    ValidityPeriod1Hour    ValidityPeriod = 0x0B  // 1 小时
    ValidityPeriod6Hours   ValidityPeriod = 0x47  // 6 小时
    ValidityPeriod24Hours  ValidityPeriod = 0xA7  // 24 小时
    ValidityPeriod1Week    ValidityPeriod = 0xAD  // 1 周
    ValidityPeriodMaximum  ValidityPeriod = 0xFF  // 最大有效期（63 周）
)
```

### PDU

PDU 数据结构。

```go
type PDU struct {
    Data   string  // PDU 数据（十六进制字符串）
    Length int     // TPDU 长度（用于 AT 命令）
}
```

### AddressType

地址类型枚举。

```go
type AddressType byte

const (
    AddressTypeUnknown       AddressType = 0x81  // 未知类型
    AddressTypeInternational AddressType = 0x91  // 国际号码
    AddressTypeAlphanumeric  AddressType = 0xD0  // 字母数字
)
```

## 编码函数

### Encode

```go
func Encode(msg *Message) ([]PDU, error)
```

将消息编码为 PDU 格式。

**参数**：

- `msg`：要编码的消息

**返回值**：

- `[]PDU`：PDU 数组（单条消息返回 1 个，长短信返回多个）
- `error`：错误信息

**特性**：

- 自动选择编码方式（如果 `msg.Encoding` 为 0）
- 自动分割长短信
- 自动生成消息引用号（如果 `msg.Reference` 为 0）

**示例**：

```go
msg := &pdu.Message{
    PhoneNumber: "+8613800138000",
    Text:        "Hello World",
    SMSC:        "+8613800138000",
}

pdus, err := pdu.Encode(msg)
if err != nil {
    log.Fatal(err)
}

for i, p := range pdus {
    fmt.Printf("PDU %d: %s\n", i+1, p.Data)
}
```

### Encode7Bit

```go
func Encode7Bit(text string) ([]byte, error)
```

将文本编码为 GSM 7-bit 格式。

**参数**：

- `text`：要编码的文本

**返回值**：

- `[]byte`：编码后的字节数组
- `error`：如果文本包含不支持的字符，返回错误

**注意**：扩展字符（€, |, ^, {, }, [, ], ~, \）会被编码为两个字节。

### EncodeUCS2

```go
func EncodeUCS2(text string) []byte
```

将文本编码为 UCS2（UTF-16 Big Endian）格式。

**参数**：

- `text`：要编码的文本

**返回值**：

- `[]byte`：编码后的字节数组

### EncodePhoneNumber

```go
func EncodePhoneNumber(number string) (AddressType, string)
```

编码电话号码为 BCD 格式。

**参数**：

- `number`：电话号码（可包含 '+' 和空格）

**返回值**：

- `AddressType`：地址类型
- `string`：编码后的十六进制字符串

## 解码函数

### Decode

```go
func Decode(pduStr string) (*Message, error)
```

解码 PDU 格式的短信。

**参数**：

- `pduStr`：PDU 十六进制字符串

**返回值**：

- `*Message`：解码后的消息
- `error`：错误信息

**支持的消息类型**：

- SMS-DELIVER（接收的短信）
- SMS-SUBMIT（发送的短信）

**示例**：

```go
pduStr := "07911326040000F0040B911346610089F60000208062917314080CC8329BFD06"
msg, err := pdu.Decode(pduStr)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("From: %s\n", msg.PhoneNumber)
fmt.Printf("Text: %s\n", msg.Text)
```

### Decode7Bit

```go
func Decode7Bit(data []byte, length int) string
```

解码 GSM 7-bit 数据。

**参数**：

- `data`：编码后的字节数组
- `length`：要解码的字符数（septets）

**返回值**：

- `string`：解码后的文本

### DecodeUCS2

```go
func DecodeUCS2(data []byte) string
```

解码 UCS2（UTF-16 Big Endian）数据。

**参数**：

- `data`：编码后的字节数组

**返回值**：

- `string`：解码后的文本

### DecodePhoneNumber

```go
func DecodePhoneNumber(data string, addrType AddressType) string
```

解码 BCD 格式的电话号码。

**参数**：

- `data`：编码后的十六进制字符串
- `addrType`：地址类型

**返回值**：

- `string`：解码后的电话号码

## 工具函数

### ValidatePhoneNumber

```go
func ValidatePhoneNumber(number string) bool
```

验证电话号码格式。

**参数**：

- `number`：电话号码

**返回值**：

- `bool`：是否有效

**规则**：

- 支持国际号码（+ 开头）和本地号码
- 长度必须在 4-15 位之间
- 只能包含数字和可选的 '+' 前缀

### IsGSM7BitCompatible

```go
func IsGSM7BitCompatible(text string) bool
```

检查文本是否兼容 GSM 7-bit 编码。

**参数**：

- `text`：要检查的文本

**返回值**：

- `bool`：是否兼容

**注意**：如果返回 `false`，应使用 UCS2 编码。

### CalculateMessageParts

```go
func CalculateMessageParts(text string, encoding Encoding) int
```

计算消息需要分割的部分数。

**参数**：

- `text`：消息文本
- `encoding`：编码类型

**返回值**：

- `int`：部分数

**长度限制**：

- GSM 7-bit：单条 160 字符，长短信每部分 153 字符
- UCS2：单条 70 字符，长短信每部分 67 字符

### GetMessageLength

```go
func GetMessageLength(text string, encoding Encoding) int
```

获取消息的实际长度。

**参数**：

- `text`：消息文本
- `encoding`：编码类型

**返回值**：

- `int`：实际长度

**注意**：对于 GSM 7-bit 编码，扩展字符计为 2 个字符。

### SwapNibbles

```go
func SwapNibbles(s string) string
```

交换字符串中每对字符的位置。

**参数**：

- `s`：输入字符串

**返回值**：

- `string`：交换后的字符串

**示例**：`"1234"` → `"2143"`

### HexToBytes

```go
func HexToBytes(hexStr string) ([]byte, error)
```

将十六进制字符串转换为字节数组。

**参数**：

- `hexStr`：十六进制字符串

**返回值**：

- `[]byte`：字节数组
- `error`：错误信息

### BytesToHex

```go
func BytesToHex(bytes []byte) string
```

将字节数组转换为十六进制字符串。

**参数**：

- `bytes`：字节数组

**返回值**：

- `string`：十六进制字符串（大写）

## 长短信管理

### ConcatManager

并发安全的长短信管理器。

```go
type ConcatManager struct { ... }
```

### NewConcatManager

```go
func NewConcatManager() *ConcatManager
```

创建新的长短信管理器。

**返回值**：

- `*ConcatManager`：管理器实例

### AddMessage

```go
func (cm *ConcatManager) AddMessage(msg *Message) (*Message, error)
```

添加一条消息。

**参数**：

- `msg`：消息

**返回值**：

- `*Message`：如果是单条消息或长短信已完整，返回完整消息；否则返回 `nil`
- `error`：错误信息

**行为**：

- 单条消息（`Parts == 0`）直接返回
- 长短信部分会被缓存，直到所有部分到齐
- 完整的长短信会自动组装并返回

**示例**：

```go
manager := pdu.NewConcatManager()

for _, pduStr := range pduStrings {
    msg, _ := pdu.Decode(pduStr)
    complete, err := manager.AddMessage(msg)
    if err != nil {
        log.Fatal(err)
    }
    if complete != nil {
        fmt.Printf("Complete message: %s\n", complete.Text)
    }
}
```

### GetPendingCount

```go
func (cm *ConcatManager) GetPendingCount() int
```

获取待完成的长短信组数。

**返回值**：

- `int`：待完成的组数

### GetPendingMessages

```go
func (cm *ConcatManager) GetPendingMessages() []*ConcatMessage
```

获取所有待完成的长短信。

**返回值**：

- `[]*ConcatMessage`：待完成的长短信数组

### Clear

```go
func (cm *ConcatManager) Clear()
```

清空所有待完成的长短信。

## 错误处理

### PDUError

自定义错误类型。

```go
type PDUError struct {
    Code    ErrorCode
    Message string
    Err     error
}
```

**方法**：

```go
func (e *PDUError) Error() string
func (e *PDUError) Unwrap() error
```

### ErrorCode

错误代码枚举。

```go
type ErrorCode int

const (
    ErrorCodeInvalidPDU          ErrorCode = 1
    ErrorCodeInvalidEncoding     ErrorCode = 2
    ErrorCodeInvalidPhoneNumber  ErrorCode = 3
    ErrorCodeMessageTooLong      ErrorCode = 4
    ErrorCodeInvalidUDH          ErrorCode = 5
)
```

### NewError

```go
func NewError(code ErrorCode, message string, err error) *PDUError
```

创建新的 PDU 错误。

**参数**：

- `code`：错误代码
- `message`：错误消息
- `err`：原始错误（可选）

**返回值**：

- `*PDUError`：错误实例

## 常量

### 长度限制

```go
const (
    MaxSingleSMSLength      = 160  // 单条短信最大长度（7-bit）
    MaxSingleSMSLengthUCS2  = 70   // 单条短信最大长度（UCS2）
    MaxConcatSMSLength      = 153  // 长短信单部分最大长度（7-bit）
    MaxConcatSMSLengthUCS2  = 67   // 长短信单部分最大长度（UCS2）
)
```

### UDH 类型

```go
const (
    UDHConcatSMS8bit  = 0x00  // 8-bit 引用的长短信 UDH
    UDHConcatSMS16bit = 0x08  // 16-bit 引用的长短信 UDH
)
```

## 性能优化

### 线程安全

- ✅ `ConcatManager`：并发安全，可在多个 goroutine 中使用
- ✅ 编码/解码函数：无状态，可并发调用
- ⚠️ `Message` 对象：非并发安全，避免在多个 goroutine 中修改同一实例

### 优化建议

1. **预分配容量**：批量处理时预先创建 `ConcatManager`
2. **对象复用**：避免频繁创建 `Message` 对象
3. **编码选择**：纯英文使用 7-bit，中文使用 UCS2
4. **错误处理**：使用类型断言检查特定错误类型
