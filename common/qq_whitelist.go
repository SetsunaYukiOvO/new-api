package common

import (
	"strings"
	"sync"
)

var qqWhitelistMutex sync.RWMutex

// IsQQWhitelisted 检查邮箱是否在QQ白名单中
// 邮箱格式为 QQ号@qq.com，提取QQ号后检查是否在白名单中
func IsQQWhitelisted(email string) bool {
	if !QQWhitelistEnabled || email == "" {
		return false
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.HasSuffix(email, "@qq.com") {
		return false
	}
	qqNumber := strings.TrimSuffix(email, "@qq.com")
	if qqNumber == "" {
		return false
	}
	qqWhitelistMutex.RLock()
	defer qqWhitelistMutex.RUnlock()
	return QQWhitelist[qqNumber]
}

// UpdateQQWhitelist 从逗号分隔的字符串更新QQ白名单
func UpdateQQWhitelist(value string) {
	qqWhitelistMutex.Lock()
	defer qqWhitelistMutex.Unlock()
	QQWhitelist = make(map[string]bool)
	if value == "" {
		return
	}
	parts := strings.Split(value, ",")
	for _, part := range parts {
		qq := strings.TrimSpace(part)
		if qq != "" {
			QQWhitelist[qq] = true
		}
	}
}

// QQWhitelistToString 将QQ白名单转为逗号分隔的字符串
func QQWhitelistToString() string {
	qqWhitelistMutex.RLock()
	defer qqWhitelistMutex.RUnlock()
	list := make([]string, 0, len(QQWhitelist))
	for qq := range QQWhitelist {
		list = append(list, qq)
	}
	return strings.Join(list, ",")
}
