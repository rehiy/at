package pdu

import "strings"

// Validator 验证器接口
type Validator interface {
	Validate() error
}

// Validate 验证消息的有效性
func (m *Message) Validate() error {
	if m.PhoneNumber == "" {
		return NewError(ErrorCodeInvalidPhoneNumber, "phone number is required", nil)
	}

	if m.Text == "" && m.Type == MessageTypeSMSSubmit {
		return NewError(ErrorCodeMessageTooLong, "message text is required", nil)
	}

	if m.Encoding != Encoding7Bit && m.Encoding != Encoding8Bit && m.Encoding != EncodingUCS2 {
		return NewError(ErrorCodeInvalidEncoding, "invalid encoding type", nil)
	}

	if m.Parts > 0 {
		if m.Part < 1 || m.Part > m.Parts {
			return NewError(ErrorCodeInvalidUDH, "invalid part number", nil)
		}
	}

	return nil
}

// ValidatePhoneNumber 验证电话号码格式
// 支持国际号码（+开头）和本地号码
// 长度必须在 4-15 位之间
func ValidatePhoneNumber(number string) bool {
	if number == "" {
		return false
	}

	cleaned := number
	if len(cleaned) > 0 && cleaned[0] == '+' {
		cleaned = cleaned[1:]
	}

	if cleaned == "" {
		return false
	}

	for i := 0; i < len(cleaned); i++ {
		if cleaned[i] < '0' || cleaned[i] > '9' {
			return false
		}
	}

	if len(cleaned) < 4 || len(cleaned) > 15 {
		return false
	}

	return true
}

// CalculateMessageParts 计算消息需要分割的部分数
// 根据编码方式和文本长度自动计算
func CalculateMessageParts(text string, encoding Encoding) int {
	var maxLen, maxConcatLen, textLen int

	if encoding == Encoding7Bit {
		maxLen = MaxSingleSMSLength
		maxConcatLen = MaxConcatSMSLength
		textLen = GetMessageLength(text, encoding)
	} else {
		maxLen = MaxSingleSMSLengthUCS2
		maxConcatLen = MaxConcatSMSLengthUCS2
		textLen = len([]rune(text))
	}

	if textLen <= maxLen {
		return 1
	}

	return (textLen + maxConcatLen - 1) / maxConcatLen
}

// GetMessageLength 获取消息的实际长度
// 对于 7-bit 编码，扩展字符计为 2 个字符
func GetMessageLength(text string, encoding Encoding) int {
	if encoding == Encoding7Bit {
		length := len([]rune(text))
		for _, r := range text {
			if _, ok := gsm7bitExtChars[r]; ok {
				length++
			}
		}
		return length
	}
	return len([]rune(text))
}

// IsGSM7BitCompatible 检查文本是否兼容 GSM 7-bit 编码
// 如果包含不支持的字符，应使用 UCS2 编码
func IsGSM7BitCompatible(text string) bool {
	for _, r := range text {
		// 检查是否在扩展字符集中
		if _, ok := gsm7bitExtChars[r]; ok {
			continue
		}
		// 检查是否在基本字符集中
		if strings.IndexRune(gsm7bitChars, r) == -1 {
			return false
		}
	}
	return true
}
