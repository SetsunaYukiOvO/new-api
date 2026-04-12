package controller

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var qqNumberRegex = regexp.MustCompile(`^\d{5,12}$`)

type qqBindCodeRequest struct {
	QQNumber string `json:"qq_number"`
}

type qqBindVerifyRequest struct {
	QQNumber string `json:"qq_number"`
	Code     string `json:"code"`
}

type uapisQQUserInfoResponse struct {
	QQ       string `json:"qq"`
	Nickname string `json:"nickname"`
	LongNick string `json:"long_nick"`
	Avatar   string `json:"avatar_url"`
	Age      int    `json:"age"`
	Sex      string `json:"sex"`
	Qid      string `json:"qid"`
	QQLevel  *int   `json:"qq_level"`
	Location string `json:"location"`
	Email    string `json:"email"`
	IsVip    bool   `json:"is_vip"`
	VipLevel int    `json:"vip_level"`
}

// QQBindGenerateCode generates a verification code for QQ binding
// User must change their QQ personal signature to this code, then call verify
func QQBindGenerateCode(c *gin.Context) {
	var req qqBindCodeRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的请求",
		})
		return
	}

	if !qqNumberRegex.MatchString(req.QQNumber) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "QQ号格式不正确",
		})
		return
	}

	// Check if this QQ is already bound to another user
	if model.IsQQIdAlreadyTaken(req.QQNumber) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该QQ号已被其他用户绑定",
		})
		return
	}

	session := sessions.Default(c)
	userId := session.Get("id").(int)

	// Check if user already has a QQ bound
	user := model.User{Id: userId}
	if err := user.FillUserById(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}
	if user.QQId != "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "你已绑定QQ号：" + user.QQId + "，请先解绑后再绑定新QQ",
		})
		return
	}

	// Generate a 6-character verification code
	code := "newapi-" + common.GetRandomString(8)

	// Store in Redis with 5 minute expiry: key = qq_bind:{userId}, value = {qqNumber}:{code}
	redisKey := fmt.Sprintf("qq_bind:%d", userId)
	redisValue := fmt.Sprintf("%s:%s", req.QQNumber, code)

	if common.RedisEnabled {
		err := common.RedisSet(redisKey, redisValue, 5*time.Minute)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "生成验证码失败，请稍后重试",
			})
			return
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "系统未启用Redis，无法使用QQ绑定功能",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"code":      code,
			"qq_number": req.QQNumber,
			"expires":   300, // 5 minutes in seconds
		},
	})
}

// QQBindVerify verifies the QQ personal signature matches the code, then binds QQ
func QQBindVerify(c *gin.Context) {
	var req qqBindVerifyRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的请求",
		})
		return
	}

	if !qqNumberRegex.MatchString(req.QQNumber) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "QQ号格式不正确",
		})
		return
	}

	session := sessions.Default(c)
	userId := session.Get("id").(int)

	// Retrieve verification code from Redis
	redisKey := fmt.Sprintf("qq_bind:%d", userId)
	if !common.RedisEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "系统未启用Redis，无法使用QQ绑定功能",
		})
		return
	}

	storedValue, err := common.RedisGet(redisKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "验证码已过期或不存在，请重新获取",
		})
		return
	}

	// Parse stored value: {qqNumber}:{code}
	expectedValue := fmt.Sprintf("%s:%s", req.QQNumber, req.Code)
	if storedValue != expectedValue {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "QQ号或验证码不匹配",
		})
		return
	}

	// Check if QQ is already bound (race condition check)
	if model.IsQQIdAlreadyTaken(req.QQNumber) {
		_ = common.RedisDel(redisKey)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该QQ号已被其他用户绑定",
		})
		return
	}

	// Call uapis.cn API to get QQ user info and check signature
	client := http.Client{Timeout: 10 * time.Second}
	apiURL := fmt.Sprintf("https://uapis.cn/api/v1/social/qq/userinfo?qq=%s", req.QQNumber)
	resp, err := client.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "查询QQ信息失败，请稍后重试",
		})
		return
	}
	defer resp.Body.Close()

	var qqInfo uapisQQUserInfoResponse
	if err := common.DecodeJson(resp.Body, &qqInfo); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "解析QQ信息失败，请稍后重试",
		})
		return
	}

	// Check if the nickname matches the verification code
	if qqInfo.Nickname != req.Code {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "验证失败：你的QQ昵称不匹配验证码，请确认已将QQ昵称修改为验证码后重试",
			"data": gin.H{
				"expected": req.Code,
				"actual":   qqInfo.Nickname,
			},
		})
		return
	}

	// Nickname matches — bind QQ to user
	if err := model.UpdateUserQQId(userId, req.QQNumber); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "绑定QQ失败，请稍后重试",
		})
		return
	}

	// Clean up Redis key
	_ = common.RedisDel(redisKey)

	// Check QQ whitelist and grant bonus quota
	bonusGranted := false
	if common.QQWhitelistEnabled && common.IsQQWhitelisted(req.QQNumber+"@qq.com") {
		_ = model.IncreaseUserQuota(userId, common.QQWhitelistQuota, true)
		model.RecordLog(userId, model.LogTypeSystem, fmt.Sprintf("QQ绑定白名单赠送 %s", logger.LogQuota(common.QQWhitelistQuota)))
		bonusGranted = true
	}

	message := "QQ绑定成功"
	if bonusGranted {
		message = fmt.Sprintf("QQ绑定成功，已赠送额度 %s", logger.LogQuota(common.QQWhitelistQuota))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data": gin.H{
			"qq_number":     req.QQNumber,
			"bonus_granted": bonusGranted,
		},
	})
}

// QQUnbind unbinds QQ from the current user
func QQUnbind(c *gin.Context) {
	if common.QQUnbindDisabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员已禁止解绑QQ",
		})
		return
	}

	session := sessions.Default(c)
	userId := session.Get("id").(int)

	user := model.User{Id: userId}
	if err := user.FillUserById(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	if user.QQId == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "你尚未绑定QQ号",
		})
		return
	}

	if err := model.UpdateUserQQId(userId, ""); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "解绑QQ失败，请稍后重试",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "QQ解绑成功",
	})
}
