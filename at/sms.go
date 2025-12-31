package at

import (
	"fmt"
	"sort"

	"github.com/rehiy/modem/pdu"
)

// SMS 短信信息
type SMS struct {
	pdu.Message
	Index  int    `json:"index"`  // SIM 卡中的索引
	Status string `json:"status"` // 状态: REC UNREAD/REC READ/STO UNSENT/STO SENT
}

// ToJSON 用于 JSON 序列化，提供前端需要的字段名
func (s *SMS) ToJSON() map[string]any {
	return map[string]any{
		"phoneNumber": s.PhoneNumber,
		"text":        s.Text,
		"time":        s.Timestamp.Format("2006/01/02 15:04:05"),
		"index":       s.Index,
		"status":      s.Status,
	}
}

// ListSMS 获取短信列表。
func (m *Device) ListSMS() ([]SMS, error) {
	responses, err := m.SendCommand("AT+CMGL=4")
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
			break // PDU 数据在下一行，如果下一行不存在，则退出
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
				Message: *sms,
				Index:   parseInt(param[0]),
				Status:  getPDUStatus(parseInt(param[1])),
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

	for _, p := range pdus {
		err = m.SendCommandExpect(fmt.Sprintf("AT+CMGS=%d", p.Length), ">")
		if err != nil {
			return err
		}

		err = m.writeString(p.Data + "\x1A")
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteSMS 删除指定索引的短信。
func (m *Device) DeleteSMS(index int) error {
	_, err := m.SendCommand(fmt.Sprintf("AT+CMGD=%d", index))
	return err
}

// 辅助函数

func getPDUStatus(stat int) string {
	switch stat {
	case 0:
		return "REC UNREAD"
	case 1:
		return "REC READ"
	case 2:
		return "STO UNSENT"
	case 3:
		return "STO SENT"
	default:
		return "UNKNOWN"
	}
}
