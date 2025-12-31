package pdu

import (
	"sync"
	"testing"
)

// BenchmarkEncode7Bit 基准测试 7-bit 编码
func BenchmarkEncode7Bit(b *testing.B) {
	text := "Hello World! This is a test message."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encode7Bit(text)
	}
}

// BenchmarkDecode7Bit 基准测试 7-bit 解码
func BenchmarkDecode7Bit(b *testing.B) {
	text := "Hello World! This is a test message."
	encoded, _ := Encode7Bit(text)
	length := GetMessageLength(text, Encoding7Bit)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Decode7Bit(encoded, length)
	}
}

// BenchmarkEncodeUCS2 基准测试 UCS2 编码
func BenchmarkEncodeUCS2(b *testing.B) {
	text := "你好世界！这是一条测试消息。"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeUCS2(text)
	}
}

// BenchmarkDecodeUCS2 基准测试 UCS2 解码
func BenchmarkDecodeUCS2(b *testing.B) {
	text := "你好世界！这是一条测试消息。"
	encoded := EncodeUCS2(text)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DecodeUCS2(encoded)
	}
}

// BenchmarkEncodeSMS 基准测试短信编码
func BenchmarkEncodeSMS(b *testing.B) {
	msg := &Message{
		PhoneNumber: "+8613800138000",
		Text:        "Hello World! This is a test message.",
		SMSC:        "+8613800138000",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encode(msg)
	}
}

// BenchmarkEncodeLongSMS 基准测试长短信编码
func BenchmarkEncodeLongSMS(b *testing.B) {
	longText := ""
	for i := 0; i < 500; i++ {
		longText += "a"
	}
	msg := &Message{
		PhoneNumber: "+8613800138000",
		Text:        longText,
		SMSC:        "+8613800138000",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encode(msg)
	}
}

// BenchmarkDecodeSMS 基准测试短信解码
func BenchmarkDecodeSMS(b *testing.B) {
	pduStr := "07911326040000F0040B911346610089F60000208062917314080CC8329BFD06"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decode(pduStr)
	}
}

// BenchmarkEncodePhoneNumber 基准测试电话号码编码
func BenchmarkEncodePhoneNumber(b *testing.B) {
	number := "+8613800138000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EncodePhoneNumber(number)
	}
}

// BenchmarkDecodePhoneNumber 基准测试电话号码解码
func BenchmarkDecodePhoneNumber(b *testing.B) {
	data := "683108108300F0"
	addrType := AddressTypeInternational
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DecodePhoneNumber(data, addrType)
	}
}

// BenchmarkConcatManager 基准测试长短信管理器
func BenchmarkConcatManager(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewConcatManager()
		for j := byte(1); j <= 3; j++ {
			msg := &Message{
				PhoneNumber: "+8613800138000",
				Text:        "Part",
				Reference:   0x42,
				Parts:       3,
				Part:        j,
			}
			_, _ = manager.AddMessage(msg)
		}
	}
}

// BenchmarkConcatManagerConcurrent 基准测试并发长短信管理
func BenchmarkConcatManagerConcurrent(b *testing.B) {
	manager := NewConcatManager()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ref := byte(0)
		for pb.Next() {
			ref++
			for j := byte(1); j <= 3; j++ {
				msg := &Message{
					PhoneNumber: "+8613800138000",
					Text:        "Part",
					Reference:   ref,
					Parts:       3,
					Part:        j,
				}
				_, _ = manager.AddMessage(msg)
			}
		}
	})
}

// BenchmarkIsGSM7BitCompatible 基准测试 GSM 7-bit 兼容性检查
func BenchmarkIsGSM7BitCompatible(b *testing.B) {
	text := "Hello World! This is a test message with some extended chars: €|^"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsGSM7BitCompatible(text)
	}
}

// BenchmarkCalculateMessageParts 基准测试消息分割计算
func BenchmarkCalculateMessageParts(b *testing.B) {
	longText := ""
	for i := 0; i < 500; i++ {
		longText += "a"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateMessageParts(longText, Encoding7Bit)
	}
}

// BenchmarkSwapNibbles 基准测试半字节交换
func BenchmarkSwapNibbles(b *testing.B) {
	s := "683108108300F0"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SwapNibbles(s)
	}
}

// BenchmarkPack7Bit 基准测试 7-bit 打包
func BenchmarkPack7Bit(b *testing.B) {
	septets := make([]byte, 160)
	for i := range septets {
		septets[i] = byte(i % 128)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pack7Bit(septets)
	}
}

// BenchmarkUnpack7Bit 基准测试 7-bit 解包
func BenchmarkUnpack7Bit(b *testing.B) {
	septets := make([]byte, 160)
	for i := range septets {
		septets[i] = byte(i % 128)
	}
	packed := pack7Bit(septets)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unpack7Bit(packed, 160)
	}
}

// BenchmarkValidatePhoneNumber 基准测试电话号码验证
func BenchmarkValidatePhoneNumber(b *testing.B) {
	number := "+8613800138000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidatePhoneNumber(number)
	}
}

// BenchmarkHexConversion 基准测试十六进制转换
func BenchmarkHexConversion(b *testing.B) {
	data := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hex := BytesToHex(data)
		_, _ = HexToBytes(hex)
	}
}

// TestConcurrentEncoding 测试并发编码
func TestConcurrentEncoding(t *testing.T) {
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := &Message{
				PhoneNumber: "+8613800138000",
				Text:        "Concurrent test message",
				SMSC:        "+8613800138000",
			}
			_, err := Encode(msg)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent encoding error: %v", err)
	}
}

// TestConcurrentDecoding 测试并发解码
func TestConcurrentDecoding(t *testing.T) {
	pduStr := "07911326040000F0040B911346610089F60000208062917314080CC8329BFD06"
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := Decode(pduStr)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent decoding error: %v", err)
	}
}
