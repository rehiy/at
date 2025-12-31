package pdu

import (
	"fmt"
	"strings"
	"sync"
)

// ConcatMessage 表示一组长短信的集合
type ConcatMessage struct {
	Reference byte              // 消息引用号
	Total     byte              // 总部分数
	Parts     map[byte]*Message // 已接收的部分
}

// NewConcatMessage 创建新的长短信集合
func NewConcatMessage(ref byte, total byte) *ConcatMessage {
	return &ConcatMessage{
		Reference: ref,
		Total:     total,
		Parts:     make(map[byte]*Message),
	}
}

// AddPart 添加一个长短信部分
func (cm *ConcatMessage) AddPart(msg *Message) error {
	if msg.Reference != cm.Reference {
		return fmt.Errorf("reference mismatch: expected %d, got %d", cm.Reference, msg.Reference)
	}
	if msg.Parts != cm.Total {
		return fmt.Errorf("total parts mismatch: expected %d, got %d", cm.Total, msg.Parts)
	}
	if msg.Part < 1 || msg.Part > cm.Total {
		return fmt.Errorf("invalid part number: %d", msg.Part)
	}

	cm.Parts[msg.Part] = msg
	return nil
}

// IsComplete 检查是否已接收所有部分
func (cm *ConcatMessage) IsComplete() bool {
	return len(cm.Parts) == int(cm.Total)
}

// GetCompleteMessage 组装完整的消息
// 按序号顺序拼接所有部分的文本
func (cm *ConcatMessage) GetCompleteMessage() (*Message, error) {
	if !cm.IsComplete() {
		return nil, fmt.Errorf("message incomplete: %d/%d parts", len(cm.Parts), cm.Total)
	}

	var completeText strings.Builder
	var firstMsg *Message

	// 预估总长度以减少内存分配
	estimatedLen := 0
	for _, part := range cm.Parts {
		estimatedLen += len(part.Text)
	}
	completeText.Grow(estimatedLen)

	for i := byte(1); i <= cm.Total; i++ {
		part, ok := cm.Parts[i]
		if !ok {
			return nil, fmt.Errorf("missing part %d", i)
		}
		if firstMsg == nil {
			firstMsg = part
		}
		completeText.WriteString(part.Text)
	}

	msg := &Message{
		Type:        firstMsg.Type,
		PhoneNumber: firstMsg.PhoneNumber,
		Text:        completeText.String(),
		Encoding:    firstMsg.Encoding,
		SMSC:        firstMsg.SMSC,
		Timestamp:   firstMsg.Timestamp,
		Flash:       firstMsg.Flash,
	}

	return msg, nil
}

// ConcatManager 管理多组长短信
type ConcatManager struct {
	mu       sync.RWMutex
	messages map[byte]*ConcatMessage
}

// NewConcatManager 创建新的长短信管理器
func NewConcatManager() *ConcatManager {
	return &ConcatManager{
		messages: make(map[byte]*ConcatMessage),
	}
}

// AddMessage 添加一条消息
func (cm *ConcatManager) AddMessage(msg *Message) (*Message, error) {
	if msg.Parts == 0 || (msg.Parts == 1 && msg.Part == 1) {
		return msg, nil
	}

	if msg.Part == 0 || msg.Part > msg.Parts {
		return nil, fmt.Errorf("invalid part number: %d/%d", msg.Part, msg.Parts)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	concat, ok := cm.messages[msg.Reference]
	if !ok {
		concat = NewConcatMessage(msg.Reference, msg.Parts)
		cm.messages[msg.Reference] = concat
	}

	err := concat.AddPart(msg)
	if err != nil {
		return nil, err
	}

	if concat.IsComplete() {
		completeMsg, err := concat.GetCompleteMessage()
		if err != nil {
			return nil, err
		}
		delete(cm.messages, msg.Reference)
		return completeMsg, nil
	}

	return nil, nil
}

// GetPendingCount 获取待完成的长短信组数
func (cm *ConcatManager) GetPendingCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.messages)
}

// GetPendingMessages 获取所有待完成的长短信
func (cm *ConcatManager) GetPendingMessages() []*ConcatMessage {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	pending := make([]*ConcatMessage, 0, len(cm.messages))
	for _, msg := range cm.messages {
		pending = append(pending, msg)
	}
	return pending
}

// Clear 清空所有待完成的长短信
func (cm *ConcatManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	clear(cm.messages)
}
