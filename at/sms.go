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
	Index       int    `json:"index"`   // 首个分片的索引
	Indices     []int  `json:"indices"` // 所有分片的索引
	Status      string `json:"status"`  // 短信状态 [PDU: TEXT, 0: "REC UNREAD", 1: "REC READ", 2: "STO UNSENT", 3: "STO SENT", 4: "ALL"]
}

// SetSMSMode 设置短信模式
// v [0: PDU 模式, 1: TEXT 模式]
func (m *Device) SetSMSMode(v int) error {
	cmd := fmt.Sprintf("%s=%d", m.commands.SMSFormat, v)
	return m.SendCommandExpect(cmd, "OK")
}

// ListSMSPdu 获取短信列表
func (m *Device) ListSMSPdu(stat int) ([]SMS, error) {
	cmd := fmt.Sprintf("%s=%d", m.commands.ListSMS, stat)
	responses, err := m.SendCommand(cmd)
	if err != nil {
		return nil, err
	}

	result := []SMS{}
	indexMap := map[byte][]int{}
	concatMgr := pdu.NewConcatManager()

	for i, l := 0, len(responses); i < l; {
		label, param := parseParam(responses[i])
		i++

		if label != "+CMGL" || len(param) < 2 {
			continue
		}

		// 无下一行，退出
		if i >= l {
			break
		}

		// 解码 PDU 数据
		msg, err := pdu.Decode(responses[i])
		i++
		if err != nil {
			m.printf("decode pdu error: %v", err)
			continue
		}

		// 记录消息索引
		index := parseInt(param[0])
		indexMap[msg.Reference] = append(indexMap[msg.Reference], index)

		// 尝试合并短信
		sms, err := concatMgr.AddMessage(msg)
		if err != nil {
			m.printf("concat sms %s error: %v", index, err)
			continue
		}

		// 添加已解析的短信到列表
		if sms != nil {
			result = append(result, SMS{
				PhoneNumber: sms.PhoneNumber,
				Text:        sms.Text,
				Time:        sms.Timestamp.Format("2006/01/02 15:04:05"),
				Index:       indexMap[sms.Reference][0],
				Indices:     indexMap[sms.Reference],
				Status:      param[1],
			})
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Index < result[j].Index })
	return result, nil
}

// SendSMSPdu 发送短信
func (m *Device) SendSMSPdu(number, message string) error {
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

// DeleteSMS 删除指定索引的短信
func (m *Device) DeleteSMS(index int) error {
	cmd := fmt.Sprintf("%s=%d", m.commands.DeleteSMS, index)
	_, err := m.SendCommand(cmd)
	return err
}
