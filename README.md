# Go Serial AT 命令通信库

一个用于与串口 modem 进行 AT 命令通信的 Go 语言库，提供完整的 AT 命令处理、响应解析和通知监听功能。

## 功能特性

- ✅ **完整的 AT 命令通信接口** - 支持同步和异步命令发送
- ✅ **自动响应解析和错误处理** - 智能识别最终响应和错误信息
- ✅ **Modem 通知监听** - 实时监听来电、短信、状态变化等通知
- ✅ **可配置的命令集** - 支持不同厂商的 modem 设备
- ✅ **短信功能** - 支持文本和 PDU 模式，自动处理中文和长短信
- ✅ **通话功能** - 拨打电话、接听、挂断等基本通话操作
- ✅ **连接管理** - 线程安全的并发操作和错误恢复机制

## 安装

```bash
go get github.com/rehiy/modem
```

## 快速开始

### 基本使用示例

```go
package main

import (
 "context"
 "log"
 "time"

 "github.com/rehiy/modem/at"
)

func main() {
 // 配置 modem 连接
 config := at.Config{
  PortName:     "/dev/ttyUSB0",
  BaudRate:     115200,
  DataBits:     8,
  StopBits:     1,
  Parity:       "N",
  ReadTimeout:  10 * time.Second,
 }

 // 创建连接
 conn, err := at.New(config)
 if err != nil {
  log.Fatal(err)
 }
 defer conn.Close()

 ctx := context.Background()

 // 测试连接
 if err := conn.Test(ctx); err != nil {
  log.Printf("Connection test failed: %v", err)
 }

 // 查询设备信息
 manufacturer, _ := conn.GetManufacturer(ctx)
 model, _ := conn.GetModel(ctx)
 log.Printf("Modem: %s %s", manufacturer, model)

 // 监听通知
 conn.ListenNotifications(ctx, func(notification string) {
  log.Printf("Notification: %s", notification)
 })
}
```

## 核心接口

### AT 接口

```go
type AT interface {
 // 基础命令操作
 SendCommand(ctx context.Context, command string) ([]string, error)
 SendCommandExpect(ctx context.Context, command string, expected string) error
 
 // 通知监听
 ListenNotifications(ctx context.Context, handler func(s string)) error
 
 // 连接管理
 Close() error
 IsConnected() bool
 
 // 设备信息查询
 Test(ctx context.Context) error
 GetManufacturer(ctx context.Context) (string, error)
 GetModel(ctx context.Context) (string, error)
 GetSignalQuality(ctx context.Context) (int, int, error)
 GetNetworkStatus(ctx context.Context) (int, int, error)
 
 // 短信功能
 SetSMSFormatText(ctx context.Context) error
 SetSMSFormatPDU(ctx context.Context) error
 SendSMSText(ctx context.Context, phoneNumber, message string) error
 SendSMSPDU(ctx context.Context, pduData string, length int) error
 ListSMS(ctx context.Context, status string) ([]SMS, error)
 ReadSMS(ctx context.Context, index int) (*SMS, error)
 DeleteSMS(ctx context.Context, index int) error
 DeleteAllSMS(ctx context.Context) error
 
 // 通话功能
 Dial(ctx context.Context, number string) error
 Answer(ctx context.Context) error
 Hangup(ctx context.Context) error
 
 // 配置管理
 Reset(ctx context.Context) error
 EchoOff(ctx context.Context) error
 EchoOn(ctx context.Context) error
}
```

## 配置说明

### Config 结构

```go
type Config struct {
 PortName        string           // 串口名称，如 '/dev/ttyUSB0' 或 'COM3'
 BaudRate        int              // 波特率，如 115200
 DataBits        int              // 数据位，如 8
 StopBits        int              // 停止位，如 1
 Parity          string           // 校验位，如 'N'（无校验）
 ReadTimeout     time.Duration    // 读取超时时间
 WriteTimeout    time.Duration    // 写入超时时间
 CommandSet      *CommandSet      // 自定义 AT 命令集
 NotificationSet *NotificationSet // 自定义通知类型集
 ResponseSet     *ResponseSet     // 自定义响应类型集
}
```

### 配置示例

```go
// 标准 modem 配置
config := at.Config{
 PortName:    "/dev/ttyUSB0",
 BaudRate:    115200,
 DataBits:    8,
 StopBits:    1,
 Parity:      "N",
 ReadTimeout: 10 * time.Second,
}

// 低速设备配置
config := at.Config{
 PortName:    "COM3",
 BaudRate:    9600,
 DataBits:    8,
 StopBits:    1,
 Parity:      "N",
 ReadTimeout: 30 * time.Second,
}
```

## 短信功能

### 短信发送示例

```go
package main

import (
 "context"
 "log"
 "time"
 "github.com/rehiy/modem/at"
)

func main() {
 config := at.Config{
  PortName:    "/dev/ttyUSB0",
  BaudRate:    115200,
  ReadTimeout: 5 * time.Second,
 }

 modem, err := at.New(config)
 if err != nil {
  log.Fatal(err)
 }
 defer modem.Close()

 ctx := context.Background()

 // 发送英文短信
 err = modem.SendSMSText(ctx, "+8613800138000", "Hello from Go!")
 if err != nil {
  log.Fatal(err)
 }
 
 // 发送中文短信（自动使用 UCS2 编码）
 err = modem.SendSMSText(ctx, "+8613800138000", "你好，这是一条中文短信！")
 if err != nil {
  log.Fatal(err)
 }
 
 // 发送长短信（自动分段）
 longMessage := "这是一条很长的中文短信，超过了70个字符的限制。" +
  "库会自动将其分割成多个段落，并作为连接短信发送。" +
  "接收者的手机会自动将这些段落重新组装成一条完整的消息。"
 err = modem.SendSMSText(ctx, "+8613800138000", longMessage)
 if err != nil {
  log.Fatal(err)
 }
 
 log.Println("短信发送完成！")
}
```

### 短信功能特性

- ✅ **自动编码检测** - 自动识别中文字符并使用 UCS2 编码
- ✅ **长短信处理** - 自动分段发送（英文 >160 字符，中文 >70 字符）
- ✅ **多模式支持** - 支持文本和 PDU 模式
- ✅ **短信管理** - 支持读取、删除、列表查询等操作

## 自定义适配

### 自定义命令集

```go
// 华为 modem 命令集
func CreateHuaweiCommandSet() at.CommandSet {
 commands := at.DefaultCommandSet()
 commands.SignalQuality = "AT^HCSQ"
 commands.ICCID = "AT^ICCID?"
 return commands
}

config := at.Config{
 PortName:    "/dev/ttyUSB0",
 BaudRate:    115200,
 CommandSet:  &huaweiCommands,
}
```

### 自定义通知集

```go
// 华为 modem 通知集
func CreateHuaweiNotificationSet() at.NotificationSet {
 notifications := at.DefaultNotificationSet()
 notifications.SignalQuality = "^HCSQ:"
 notifications.NetworkReg = "^CREG:"
 return notifications
}

config := at.Config{
 PortName:        "/dev/ttyUSB0",
 BaudRate:        115200,
 NotificationSet: &huaweiNotifications,
}
```

## 错误处理

```go
// 检查错误类型
if at.IsConnectionError(err) {
 // 处理连接错误
} else if at.IsCommandError(err) {
 // 处理命令错误
}

// 获取详细错误信息
if atErr, ok := err.(*at.ATError); ok {
 log.Printf("Command %q failed with response: %v", atErr.Command, atErr.Response)
}
```

## 依赖

- [github.com/tarm/serial](https://github.com/tarm/serial) - 串口通信库

## 许可证

MIT License
