# 使用示例

本文档提供了 rehiy-modem-pdu 库的详细使用示例和最佳实践。

## 🎯 快速导航

- [基本使用](#基本使用)
- [长短信处理](#长短信处理)
- [中文和特殊字符](#中文和特殊字符)
- [错误处理](#错误处理)
- [并发使用](#并发使用)
- [最佳实践](#最佳实践)

## 基本使用

### 编码短信

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    // 创建消息
    msg := &pdu.Message{
        PhoneNumber: "+8613800138000",
        Text:        "Hello World!",
        SMSC:        "+8613800138000",
    }

    // 编码为 PDU
    pdus, err := pdu.Encode(msg)
    if err != nil {
        panic(err)
    }

    // 输出 PDU 数据
    for i, p := range pdus {
        fmt.Printf("PDU %d: %s (Length: %d)\n", i+1, p.Data, p.Length)
    }
}
```

### 解码短信

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    // PDU 字符串
    pduStr := "07911326040000F0040B911346610089F60000208062917314080CC8329BFD06"

    // 解码
    msg, err := pdu.Decode(pduStr)
    if err != nil {
        panic(err)
    }

    // 输出消息内容
    fmt.Printf("From: %s\n", msg.PhoneNumber)
    fmt.Printf("Text: %s\n", msg.Text)
    fmt.Printf("Time: %s\n", msg.Timestamp)
}
```

## 长短信处理

### 发送长短信

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    // 创建长消息（超过 160 字符）
    longText := "This is a very long message that will be automatically split into multiple parts..."
    
    msg := &pdu.Message{
        PhoneNumber: "+8613800138000",
        Text:        longText,
        SMSC:        "+8613800138000",
    }

    // 自动分割为多个 PDU
    pdus, err := pdu.Encode(msg)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Message split into %d parts\n", len(pdus))

    // 发送每个 PDU
    for i, p := range pdus {
        fmt.Printf("Sending part %d: %s\n", i+1, p.Data)
        // 这里调用调制解调器发送 PDU
    }
}
```

### 接收长短信

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    // 创建长短信管理器
    manager := pdu.NewConcatManager()
    
    // 模拟接收多个 PDU（长短信的各个部分）
    pduStrings := []string{
        "0500030C02010007911326040000F0040B911346610089F60000208062917314080CC8329BFD06",
        "0500030C02020007911326040000F0040B911346610089F60000208062917314080CC8329BFD06",
    }
    
    for _, pduStr := range pduStrings {
        // 解码 PDU
        msg, err := pdu.Decode(pduStr)
        if err != nil {
            fmt.Printf("Error decoding PDU: %v\n", err)
            continue
        }
        
        // 添加到长短信管理器
        complete, err := manager.AddMessage(msg)
        if err != nil {
            fmt.Printf("Error adding message: %v\n", err)
            continue
        }
        
        if complete != nil {
            // 处理完整消息
            fmt.Printf("Complete message received: %s\n", complete.Text)
        }
    }
}
```

## 中文和特殊字符

### 中文短信（UCS2 编码）

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    msg := &pdu.Message{
        PhoneNumber: "+8613800138000",
        Text:        "你好世界！",
        SMSC:        "+8613800138000",
        Encoding:    pdu.EncodingUCS2, // 指定 UCS2 编码
    }

    pdus, err := pdu.Encode(msg)
    if err != nil {
        panic(err)
    }

    fmt.Printf("PDU: %s\n", pdus[0].Data)
}
```

### 自动编码选择

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    // 英文文本，自动选择 7-bit 编码
    msg1 := &pdu.Message{
        PhoneNumber: "+8613800138000",
        Text:        "Hello World",
        SMSC:        "+8613800138000",
        // 不设置 Encoding，让库自动选择
    }
    
    // 中文文本，自动选择 UCS2 编码
    msg2 := &pdu.Message{
        PhoneNumber: "+8613800138000",
        Text:        "你好世界",
        SMSC:        "+8613800138000",
        // 不设置 Encoding，让库自动选择
    }
    
    pdus1, _ := pdu.Encode(msg1)
    pdus2, _ := pdu.Encode(msg2)
    
    fmt.Printf("English PDU: %s\n", pdus1[0].Data)
    fmt.Printf("Chinese PDU: %s\n", pdus2[0].Data)
}
```

### GSM 7-bit 扩展字符

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    // 包含扩展字符的文本
    msg := &pdu.Message{
        PhoneNumber: "+8613800138000",
        Text:        "Price: €10 [test] {data} a|b",
        SMSC:        "+8613800138000",
        Encoding:    pdu.Encoding7Bit, // 使用 7-bit 编码
    }

    pdus, err := pdu.Encode(msg)
    if err != nil {
        panic(err)
    }

    fmt.Printf("PDU with extended chars: %s\n", pdus[0].Data)
    
    // 解码验证
    decoded, _ := pdu.Decode(pdus[0].Data)
    fmt.Printf("Decoded text: %s\n", decoded.Text)
}
```

## 错误处理

### 基本错误处理

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    // 编码错误处理
    msg := &pdu.Message{
        PhoneNumber: "invalid-number", // 无效号码
        Text:        "Hello",
        SMSC:        "+8613800138000",
    }
    
    _, err := pdu.Encode(msg)
    if err != nil {
        fmt.Printf("Encoding error: %v\n", err)
    }
    
    // 解码错误处理
    _, err = pdu.Decode("invalid-pdu-string")
    if err != nil {
        fmt.Printf("Decoding error: %v\n", err)
    }
}
```

### 特定错误类型检查

```go
package main

import (
    "fmt"
    "github.com/rehiy/modem/pdu"
)

func main() {
    msg, err := pdu.Decode("invalid-pdu")
    if err != nil {
        // 检查特定错误类型
        if pduErr, ok := err.(*pdu.PDUError); ok {
            switch pduErr.Code {
            case pdu.ErrorCodeInvalidPDU:
                fmt.Println("Invalid PDU format")
            case pdu.ErrorCodeInvalidEncoding:
                fmt.Println("Unsupported encoding")
            case pdu.ErrorCodeInvalidPhoneNumber:
                fmt.Println("Invalid phone number")
            default:
                fmt.Printf("PDU error: %v\n", err)
            }
        } else {
            fmt.Printf("Other error: %v\n", err)
        }
    } else {
        fmt.Printf("Decoded message: %s\n", msg.Text)
    }
}
```

## 并发使用

### 并发安全的长短信处理

```go
package main

import (
    "fmt"
    "sync"
    "github.com/rehiy/modem/pdu"
)

func main() {
    manager := pdu.NewConcatManager()
    var wg sync.WaitGroup
    
    // 模拟多个并发接收的 PDU
    pduStrings := []string{
        "0500030C02010007911326040000F0040B911346610089F60000208062917314080CC8329BFD06",
        "0500030C02020007911326040000F0040B911346610089F60000208062917314080CC8329BFD06",
    }
    
    for i, pduStr := range pduStrings {
        wg.Add(1)
        go func(index int, pduData string) {
            defer wg.Done()
            
            msg, err := pdu.Decode(pduData)
            if err != nil {
                fmt.Printf("Goroutine %d: decoding error: %v\n", index, err)
                return
            }
            
            complete, err := manager.AddMessage(msg)
            if err != nil {
                fmt.Printf("Goroutine %d: add message error: %v\n", index, err)
                return
            }
            
            if complete != nil {
                fmt.Printf("Goroutine %d: complete message: %s\n", index, complete.Text)
            }
        }(i, pduStr)
    }
    
    wg.Wait()
    fmt.Printf("Pending messages: %d\n", manager.GetPendingCount())
}
```

## 最佳实践

### 1. 自动编码选择

```go
// ✅ 推荐：让库自动选择最优编码
msg := &pdu.Message{
    PhoneNumber: "+8613800138000",
    Text:        "Hello 世界",  // 包含中文，自动选择 UCS2
    SMSC:        "+8613800138000",
    // 不设置 Encoding，让库自动选择
}

// ❌ 不推荐：手动指定编码（除非有特殊需求）
msg2 := &pdu.Message{
    PhoneNumber: "+8613800138000",
    Text:        "Hello 世界",
    SMSC:        "+8613800138000",
    Encoding:    pdu.Encoding7Bit, // 可能导致编码错误
}
```

### 2. 输入验证

```go
msg := &pdu.Message{
    PhoneNumber: "+8613800138000",
    Text:        "Hello",
    SMSC:        "+8613800138000",
}

// 编码前验证消息
if err := msg.Validate(); err != nil {
    log.Fatal(err)
}

pdus, err := pdu.Encode(msg)
if err != nil {
    log.Fatal(err)
}
```

### 3. 长短信管理器复用

```go
// 创建全局的长短信管理器
var globalManager = pdu.NewConcatManager()

func processPDU(pduStr string) {
    msg, err := pdu.Decode(pduStr)
    if err != nil {
        return
    }
    
    complete, err := globalManager.AddMessage(msg)
    if err != nil {
        return
    }
    
    if complete != nil {
        // 处理完整消息
        handleCompleteMessage(complete)
    }
}
```

### 4. 性能优化

```go
// 批量处理时复用对象
func processBatch(pduStrings []string) {
    manager := pdu.NewConcatManager()
    
    for _, pduStr := range pduStrings {
        msg, _ := pdu.Decode(pduStr)
        complete, _ := manager.AddMessage(msg)
        if complete != nil {
            // 批量处理完整消息
        }
    }
    
    // 清理未完成的消息
    manager.Clear()
}
```

## 常见问题

### Q: 如何判断文本是否兼容 GSM 7-bit 编码？

```go
if pdu.IsGSM7BitCompatible("Hello World") {
    fmt.Println("文本兼容 GSM 7-bit 编码")
} else {
    fmt.Println("文本需要 UCS2 编码")
}
```

### Q: 如何计算消息需要分割的部分数？

```go
parts := pdu.CalculateMessageParts("长文本内容", pdu.Encoding7Bit)
fmt.Printf("消息需要分割为 %d 部分\n", parts)
```

### Q: 如何获取消息的实际长度？

```go
length := pdu.GetMessageLength("文本内容", pdu.Encoding7Bit)
fmt.Printf("消息实际长度：%d 字符\n", length)
```
