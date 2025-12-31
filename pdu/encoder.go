package pdu

import (
	"fmt"
	"strings"
	"time"
)

// Encode 将消息编码为 PDU 格式
// 自动选择编码方式，并根据长度决定是否分割为长短信
func Encode(msg *Message) ([]PDU, error) {
	encoding := msg.Encoding
	if encoding == 0 {
		if !IsGSM7BitCompatible(msg.Text) {
			encoding = EncodingUCS2
		} else {
			encoding = Encoding7Bit
		}
	}

	var maxLen, maxConcatLen, textLen int

	if encoding == Encoding7Bit {
		maxLen = MaxSingleSMSLength
		maxConcatLen = MaxConcatSMSLength
		textLen = GetMessageLength(msg.Text, encoding)
	} else {
		maxLen = MaxSingleSMSLengthUCS2
		maxConcatLen = MaxConcatSMSLengthUCS2
		textLen = len([]rune(msg.Text))
	}

	if textLen <= maxLen {
		pdu, err := encodeSingle(msg, encoding)
		if err != nil {
			return nil, err
		}
		return []PDU{pdu}, nil
	}

	return encodeConcat(msg, encoding, maxConcatLen)
}

// encodeSingle 编码单条短信
func encodeSingle(msg *Message, encoding Encoding) (PDU, error) {
	var pdu strings.Builder
	pdu.Grow(256)

	smscHex, err := encodeSMSC(msg.SMSC)
	if err != nil {
		return PDU{}, err
	}
	pdu.WriteString(smscHex)

	// PDU-Type: 0x01=SMS-SUBMIT, 0x20=状态报告请求, 0x10=有效期, 0x40=包含UDH
	pduType := byte(0x01)
	if msg.RequestReport {
		pduType |= 0x20
	}
	if msg.ValidityPeriod != 0 {
		pduType |= 0x10
	}
	if len(msg.UDH) > 0 {
		pduType |= 0x40
	}
	pdu.WriteString(fmt.Sprintf("%02X", pduType))

	// Message Reference (00 表示由设备自动分配)
	pdu.WriteString("00")

	// 编码目标地址
	addrType, addrHex := EncodePhoneNumber(msg.PhoneNumber)
	// 计算数字长度（不包括 '+' 和 'F' 填充）
	addrLen := 0
	for i := 0; i < len(msg.PhoneNumber); i++ {
		if msg.PhoneNumber[i] >= '0' && msg.PhoneNumber[i] <= '9' {
			addrLen++
		}
	}
	pdu.WriteString(fmt.Sprintf("%02X%02X%s", addrLen, addrType, addrHex))

	// Protocol Identifier (00 表示普通短信)
	pdu.WriteString("00")

	// Data Coding Scheme (编码方式)
	dcs := byte(encoding)
	if msg.Flash {
		dcs |= 0x10
	}
	pdu.WriteString(fmt.Sprintf("%02X", dcs))

	if msg.ValidityPeriod != 0 {
		pdu.WriteString(fmt.Sprintf("%02X", msg.ValidityPeriod))
	}

	userData, udl, err := encodeUserData(msg.Text, encoding, msg.UDH)
	if err != nil {
		return PDU{}, err
	}
	pdu.WriteString(fmt.Sprintf("%02X%s", udl, userData))

	pduStr := pdu.String()
	smscLen := len(smscHex) / 2
	tpduLen := len(pduStr)/2 - smscLen

	return PDU{
		Data:   pduStr,
		Length: tpduLen,
	}, nil
}

// encodeConcat 编码长短信（多部分消息）
func encodeConcat(msg *Message, encoding Encoding, maxLen int) ([]PDU, error) {
	text := msg.Text
	runes := []rune(text)

	var parts []string
	if encoding == Encoding7Bit {
		parts = splitText7Bit(text, maxLen)
	} else {
		estimatedParts := (len(runes) + maxLen - 1) / maxLen
		parts = make([]string, 0, estimatedParts)
		for i := 0; i < len(runes); i += maxLen {
			end := i + maxLen
			if end > len(runes) {
				end = len(runes)
			}
			parts = append(parts, string(runes[i:end]))
		}
	}

	totalParts := byte(len(parts))
	reference := msg.Reference
	if reference == 0 {
		reference = byte(time.Now().Unix() & 0xFF)
	}

	pdus := make([]PDU, 0, len(parts))
	for i, part := range parts {
		partMsg := *msg
		partMsg.Text = part
		partMsg.Reference = reference
		partMsg.Parts = totalParts
		partMsg.Part = byte(i + 1)

		udh := []byte{
			0x00, 0x03, reference, totalParts, byte(i + 1),
		}
		partMsg.UDH = udh

		pdu, err := encodeSingle(&partMsg, encoding)
		if err != nil {
			return nil, err
		}
		pdus = append(pdus, pdu)
	}

	return pdus, nil
}

// encodeSMSC 编码短信中心号码
func encodeSMSC(smsc string) (string, error) {
	if smsc == "" {
		return "00", nil
	}

	addrType, addrHex := EncodePhoneNumber(smsc)
	length := len(addrHex)/2 + 1

	return fmt.Sprintf("%02X%02X%s", length, addrType, addrHex), nil
}

// encodeUserData 编码用户数据（包括 UDH 和文本）
// 返回十六进制字符串、UDL（用户数据长度）和错误
func encodeUserData(text string, encoding Encoding, udh []byte) (string, int, error) {
	var userData []byte
	var udl int

	if len(udh) > 0 {
		udhLen := byte(len(udh))
		userData = append(userData, udhLen)
		userData = append(userData, udh...)
	}

	var textData []byte
	var err error

	switch encoding {
	case Encoding7Bit:
		textData, err = Encode7Bit(text)
		if err != nil {
			return "", 0, err
		}
		textLen := GetMessageLength(text, encoding)
		if len(udh) > 0 {
			// 计算填充位：UDH 后需要对齐到 7-bit 边界
			udhBits := (len(udh) + 1) * 8
			padding := 7 - (udhBits % 7)
			if padding == 7 {
				padding = 0
			}
			udl = textLen + ((udhBits + padding) / 7)
			textData = shiftLeft(textData, padding)
		} else {
			udl = textLen
		}
	case Encoding8Bit:
		textData = []byte(text)
		udl = len(userData) + len(textData)
	case EncodingUCS2:
		textData = EncodeUCS2(text)
		udl = len(userData) + len(textData)
	default:
		return "", 0, fmt.Errorf("unsupported encoding: %d", encoding)
	}

	userData = append(userData, textData...)
	return BytesToHex(userData), udl, nil
}

// splitText7Bit 按 GSM 7-bit 字符长度分割文本
// 扩展字符（如 €）占用 2 个字符位置
func splitText7Bit(text string, maxLen int) []string {
	parts := make([]string, 0, (len(text)/maxLen)+1)
	var current strings.Builder
	current.Grow(maxLen)
	currentLen := 0

	for _, r := range text {
		charLen := 1
		if _, ok := gsm7bitExtChars[r]; ok {
			charLen = 2
		}

		if currentLen+charLen > maxLen {
			parts = append(parts, current.String())
			current.Reset()
			current.Grow(maxLen)
			currentLen = 0
		}

		current.WriteRune(r)
		currentLen += charLen
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// shiftLeft 将字节数组左移指定位数
// 用于 7-bit 编码中的填充对齐
func shiftLeft(data []byte, bits int) []byte {
	if bits == 0 || len(data) == 0 {
		return data
	}

	result := make([]byte, len(data))
	carry := byte(0)

	for i := len(data) - 1; i >= 0; i-- {
		result[i] = (data[i] << bits) | carry
		carry = data[i] >> (8 - bits)
	}

	return result
}
