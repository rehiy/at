package at

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rehiy/modem/pdu"
)

type SMS struct {
	Index   int    `json:"index"`
	Status  string `json:"status"`
	Number  string `json:"number"`
	Time    string `json:"time"`
	Message string `json:"message"`
}

// ListSMS 获取短信列表。
func (s *Device) ListSMS() ([]SMS, error) {
	responses, err := s.SendCommand("AT+CMGL=4")
	if err != nil {
		return nil, err
	}

	// 按引用号+号码分组存储长短信片段
	type fragment struct {
		seq   int
		sms   SMS
		ref   int
		total int
	}
	fragments := make(map[string][]fragment)
	var singles []SMS

	for i := 0; i < len(responses); i++ {
		label, param := parseParam(responses[i])
		if label != "+CMGL" || len(param) < 2 {
			continue
		}

		// PDU 数据在下一行
		if i+1 >= len(responses) {
			continue
		}
		pduHex := strings.TrimSpace(strings.TrimSuffix(responses[i+1], "OK"))

		idx := parseInt(param[0])
		stat := parseInt(param[1])
		item := SMS{Index: idx, Status: getPDUStatus(stat)}
		ref, total, seq := 0, 1, 1

		msg, err := pdu.Decode(pduHex)
		if err != nil {
			item.Message = "PDU Decode Error: " + err.Error()
		} else {
			item.Number = msg.PhoneNumber
			item.Time = msg.Timestamp.Format("2006/01/02 15:04:05")
			item.Message = msg.Text
			if msg.Parts > 0 {
				total = int(msg.Parts)
				seq = int(msg.Part)
				ref = int(msg.Reference)
			}
		}

		if total > 1 {
			key := fmt.Sprintf("%s_%d", item.Number, ref)
			fragments[key] = append(fragments[key], fragment{seq: seq, sms: item})
		} else {
			singles = append(singles, item)
		}
	}

	// 合并长短信
	for _, frags := range fragments {
		sort.Slice(frags, func(i, j int) bool { return frags[i].seq < frags[j].seq })
		var fullMsg string
		for _, f := range frags {
			fullMsg += f.sms.Message
		}
		frags[0].sms.Message = fullMsg
		singles = append(singles, frags[0].sms)
	}

	sort.Slice(singles, func(i, j int) bool { return singles[i].Index < singles[j].Index })
	return singles, nil
}

// SendSMS 发送短信。
func (s *Device) SendSMS(number, message string) error {
	msg := &pdu.Message{
		Type:        pdu.MessageTypeSMSSubmit,
		PhoneNumber: number,
		Text:        message,
	}

	if !pdu.IsGSM7BitCompatible(message) {
		msg.Encoding = pdu.EncodingUCS2
	} else {
		msg.Encoding = pdu.Encoding7Bit
	}

	pdus, err := pdu.Encode(msg)
	if err != nil {
		return err
	}

	for _, p := range pdus {
		err = s.SendCommandExpect(fmt.Sprintf("AT+CMGS=%d", p.Length), ">")
		if err != nil {
			return err
		}

		err = s.writeString(p.Data + "\x1A")
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteSMS 删除指定索引的短信。
func (s *Device) DeleteSMS(index int) error {
	_, err := s.SendCommand(fmt.Sprintf("AT+CMGD=%d", index))
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
