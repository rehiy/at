package at

import (
	"fmt"
	"strings"
)

// NotificationSet 定义可配置的URC（Unsolicited Result Code）类型集合
type NotificationSet struct {
	Ring            string // 来电通知（URC）
	SMSReady        string // 新短信通知（URC）
	SMSContent      string // 短信内容通知（URC）
	SMSStatusReport string // 短信状态报告（URC）
	CellBroadcast   string // 小区广播（URC）
	CallRing        string // 来电铃声（URC）
	CallerID        string // 来电显示（URC）
	CallWaiting     string // 呼叫等待（URC）
	NetworkReg      string // 网络注册状态（URC）
	GPRSReg         string // GPRS 注册状态（URC）
	EPSReg          string // EPS 注册状态（URC）
	USSD            string // USSD 响应（URC）
	StatusChange    string // 状态变化通知（URC）
}

// DefaultNotificationSet 返回默认的URC类型集合
func DefaultNotificationSet() NotificationSet {
	return NotificationSet{
		Ring:            "RING",
		SMSReady:        "+CMTI:",
		SMSContent:      "+CMT:",
		SMSStatusReport: "+CDS:",
		CellBroadcast:   "+CBM:",
		CallRing:        "+CRING:",
		CallerID:        "+CLIP:",
		CallWaiting:     "+CCWA:",
		NetworkReg:      "+CREG:",
		GPRSReg:         "+CGREG:",
		EPSReg:          "+CEREG:",
		USSD:            "+CUSD:",
		StatusChange:    "+CIEV:",
	}
}

// GetAllNotifications 返回所有URC前缀的列表
func (ns *NotificationSet) GetAllNotifications() []string {
	return []string{
		ns.Ring,
		ns.SMSReady,
		ns.SMSContent,
		ns.SMSStatusReport,
		ns.CellBroadcast,
		ns.CallRing,
		ns.CallerID,
		ns.CallWaiting,
		ns.NetworkReg,
		ns.GPRSReg,
		ns.EPSReg,
		ns.USSD,
		ns.StatusChange,
	}
}

// IsNotification 检查是否为通知消息
func (ns *NotificationSet) IsNotification(line string) bool {
	for _, notification := range ns.GetAllNotifications() {
		if notification != "" && strings.HasPrefix(line, notification) {
			return true
		}
	}
	return false
}

// ParseNotification 解析通知内容
func ParseNotification(notification string) (string, map[string]string) {
	if strings.Contains(notification, ":") {
		parts := strings.SplitN(notification, ":", 2)
		if len(parts) == 2 {
			noteType := strings.TrimSpace(parts[0])
			params := parseNotificationParams(parts[1])
			return noteType, params
		}
	}

	return notification, nil
}

// parseNotificationParams 解析通知参数
func parseNotificationParams(paramStr string) map[string]string {
	params := make(map[string]string)

	// 移除引号并分割参数
	parts := strings.Split(trimQuotes(paramStr), ",")

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			params[fmt.Sprintf("param%d", i+1)] = part
		}
	}

	return params
}
