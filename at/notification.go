package at

import (
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
func DefaultNotificationSet() *NotificationSet {
	return &NotificationSet{
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
