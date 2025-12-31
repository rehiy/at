package at

import (
	"fmt"
	"sort"
	"time"

	"github.com/rehiy/modem/pdu"
)

// SMS 短信信息
type SMS struct {
	PhoneNumber string `json:"phoneNumber"`
	Text        string `json:"text"`
	Time        string `json:"time"`
	Index       int    `json:"index"`  // SIM 卡中的索引
	Status      int    `json:"status"` // 0: 未读, 1: 已读, 2: 未发送, 3: 已发送
}

// SetSMSTextMode 设置短信为文本模式
func (m *Device) SetSMSTextMode() error {
	cmd := fmt.Sprintf("%s=0", m.commands.SMSFormat)
	return m.SendCommandExpect(cmd, "OK")
}

// SetSMSPDUMode 设置短信为PDU模式
func (m *Device) SetSMSPDUMode() error {
	cmd := fmt.Sprintf("%s=1", m.commands.SMSFormat)
	return m.SendCommandExpect(cmd, "OK")
}

// ListSMS 获取短信列表。
func (m *Device) ListSMS() ([]SMS, error) {
	cmd := fmt.Sprintf("%s=%d", m.commands.ListSMS, 4)
	responses, err := m.SendCommand(cmd)
	if err != nil {
		return nil, err
	}

	var result []SMS
	concatMgr := pdu.NewConcatManager()

	for i, l := 0, len(responses); i < l; {
		label, param := parseParam(responses[i])
		i++

		if label != "+CMGL" || len(param) < 2 {
			continue
		}

		if i >= l {
			break // 下一行找不到 PDU 数据，退出
		}

		pduHex := responses[i]
		i++

		msg, err := pdu.Decode(pduHex)
		if err != nil {
			m.printf("decode pdu error: %v", err)
			continue
		}

		sms, err := concatMgr.AddMessage(msg)
		if err != nil {
			m.printf("concat sms %s error: %v", param[0], err)
			continue
		}

		if sms != nil {
			result = append(result, SMS{
				PhoneNumber: sms.PhoneNumber,
				Text:        sms.Text,
				Time:        sms.Timestamp.Format("2006/01/02 15:04:05"),
				Index:       parseInt(param[0]),
				Status:      parseInt(param[1]),
			})
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Index < result[j].Index })
	return result, nil
}

// SendSMS 发送短信。
func (m *Device) SendSMS(number, message string) error {
	msg := &pdu.Message{
		Type:        pdu.MessageTypeSMSSubmit,
		PhoneNumber: number,
		Text:        message,
	}

	pdus, err := pdu.Encode(msg)
	if err != nil {
		return err
	}

	timeout := m.timeout
	m.timeout = time.Second * 15

	for _, p := range pdus {
		cmd := fmt.Sprintf("%s=%d", m.commands.SendSMS, p.Length)
		err = m.SendCommandExpect(cmd, ">")
		if err != nil {
			m.printf("send sms command error: %v", err)
			return err
		}

		_, err = m.SendCommand(p.Data + "\x1A")
		if err != nil {
			m.printf("send sms response error: %v", err)
			return err
		}
	}

	m.timeout = timeout
	return nil
}

// DeleteSMS 删除指定索引的短信。
func (m *Device) DeleteSMS(index int) error {
	cmd := fmt.Sprintf("%s=%d", m.commands.DeleteSMS, index)
	_, err := m.SendCommand(cmd)
	return err
}

// 辅助函数
