package service

import "github.com/gin-gonic/gin"

const errorPassthroughServiceContextKey = "error_passthrough_service"

// BindErrorPassthroughService 将错误透传服务绑定到请求上下文，供 service 层在非 failover 场景下复用规则。
func BindErrorPassthroughService(c *gin.Context, svc *ErrorPassthroughService) {
	if c == nil || svc == nil {
		return
	}
	c.Set(errorPassthroughServiceContextKey, svc)
}

func getBoundErrorPassthroughService(c *gin.Context) *ErrorPassthroughService {
	if c == nil {
		return nil
	}
	v, ok := c.Get(errorPassthroughServiceContextKey)
	if !ok {
		return nil
	}
	svc, ok := v.(*ErrorPassthroughService)
	if !ok {
		return nil
	}
	return svc
}

type passthroughRuleResult struct {
	StatusCode  int
	ErrType     string
	ErrMessage  string
	ErrCode     any
	HasErrCode  bool
	ErrParam    any
	HasErrParam bool
	ErrStatus   string
}

func (r passthroughRuleResult) anthropicPayload() gin.H {
	errObj := gin.H{
		"type":    r.ErrType,
		"message": r.ErrMessage,
	}
	if r.HasErrCode {
		errObj["code"] = r.ErrCode
	}
	if r.HasErrParam {
		errObj["param"] = r.ErrParam
	}
	return gin.H{
		"type":  "error",
		"error": errObj,
	}
}

func (r passthroughRuleResult) openAIPayload() gin.H {
	errObj := gin.H{
		"type":    r.ErrType,
		"message": r.ErrMessage,
	}
	if r.HasErrCode {
		errObj["code"] = r.ErrCode
	}
	if r.HasErrParam {
		errObj["param"] = r.ErrParam
	}
	return gin.H{"error": errObj}
}

func (r passthroughRuleResult) geminiPayload() gin.H {
	errObj := gin.H{
		"message": r.ErrMessage,
	}
	if r.HasErrCode {
		errObj["code"] = r.ErrCode
	} else {
		errObj["code"] = r.StatusCode
	}
	if r.ErrStatus != "" {
		errObj["status"] = r.ErrStatus
	}
	return gin.H{
		"type":  "error",
		"error": errObj,
	}
}

// applyErrorPassthroughRule 按规则改写错误响应；未命中时返回默认响应参数。
func applyErrorPassthroughRule(
	c *gin.Context,
	platform string,
	upstreamStatus int,
	responseBody []byte,
	defaultStatus int,
	defaultErrType string,
	defaultErrMsg string,
) (passthroughRuleResult, bool) {
	result := passthroughRuleResult{
		StatusCode: defaultStatus,
		ErrType:    defaultErrType,
		ErrMessage: defaultErrMsg,
	}

	svc := getBoundErrorPassthroughService(c)
	if svc == nil {
		return result, false
	}

	rule := svc.MatchRule(platform, upstreamStatus, responseBody)
	if rule == nil {
		return result, false
	}

	info := extractUpstreamErrorInfo(responseBody)
	result.StatusCode = upstreamStatus
	if !rule.PassthroughCode && rule.ResponseCode != nil {
		result.StatusCode = *rule.ResponseCode
	}
	if result.StatusCode <= 0 {
		result.StatusCode = upstreamStatus
	}

	if info.HasCode {
		result.ErrCode = info.Code
		result.HasErrCode = true
	}
	if info.HasParam {
		result.ErrParam = info.Param
		result.HasErrParam = true
	}

	if rule.PassthroughBody {
		if info.Type != "" {
			result.ErrType = info.Type
		}
		if info.Message != "" {
			result.ErrMessage = info.Message
		}
		result.ErrStatus = info.Status
	} else if rule.CustomMessage != nil {
		result.ErrMessage = *rule.CustomMessage
	}

	if result.ErrType == "" {
		result.ErrType = "upstream_error"
	}
	if result.ErrMessage == "" {
		result.ErrMessage = defaultErrMsg
	}

	// 命中 skip_monitoring 时在 context 中标记，供 ops_error_logger 跳过记录。
	if rule.SkipMonitoring {
		c.Set(OpsSkipPassthroughKey, true)
	}

	return result, true
}
