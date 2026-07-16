package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/thanhpk/randstr"
)

type BepusdtPayRequest struct {
	Amount int64 `json:"amount"`
}

type bepusdtCreateResp struct {
	StatusCode int64  `json:"status_code"`
	Message    string `json:"message"`
	Data       struct {
		TradeID    string `json:"trade_id"`
		PaymentURL string `json:"payment_url"`
	} `json:"data"`
}

func getBepusdtPayMoney(amount float64, group string) float64 {
	originalAmount := amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		amount = amount / common.QuotaPerUnit
	}
	topupGroupRatio := common.GetTopupGroupRatio(group)
	if topupGroupRatio == 0 {
		topupGroupRatio = 1
	}
	discount := 1.0
	if ds, ok := operation_setting.GetPaymentSetting().AmountDiscount[int(originalAmount)]; ok {
		if ds > 0 {
			discount = ds
		}
	}
	return amount * setting.BepusdtUnitPrice * topupGroupRatio * discount
}

func getBepusdtFiat() string {
	fiat := strings.TrimSpace(setting.BepusdtFiat)
	if fiat == "" {
		return "USD"
	}
	return fiat
}

// RequestBepusdtAmount returns the charged money for a topup amount.
func RequestBepusdtAmount(c *gin.Context) {
	var req BepusdtPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	minTopup := int64(setting.BepusdtMinTopUp)
	if req.Amount < minTopup {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", minTopup)})
		return
	}
	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}
	payMoney := getBepusdtPayMoney(float64(req.Amount), group)
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": strconv.FormatFloat(payMoney, 'f', 2, 64)})
}

// RequestBepusdtPay creates a BEPUSDT payment order and returns the payment URL.
func RequestBepusdtPay(c *gin.Context) {
	if !isBepusdtTopUpEnabled() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "BEPUSDT 支付未启用"})
		return
	}

	var req BepusdtPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	minTopup := int64(setting.BepusdtMinTopUp)
	if req.Amount < minTopup {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", minTopup)})
		return
	}

	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}

	payMoney := getBepusdtPayMoney(float64(req.Amount), group)
	if payMoney < 0.01 {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}

	// Token 模式下归一化 Amount（存等价美元/CNY 数量，避免 Recharge 双重放大）
	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		amount = int64(float64(req.Amount) / common.QuotaPerUnit)
		if amount < 1 {
			amount = 1
		}
	}

	tradeNo := fmt.Sprintf("BEPUSDT-%d-%d-%s", id, time.Now().UnixMilli(), randstr.String(6))
	topUp := &model.TopUp{
		UserId:          id,
		Amount:          amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodBepusdt,
		PaymentProvider: model.PaymentProviderBepusdt,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("BEPUSDT 创建充值订单失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	callbackAddr := service.GetCallbackAddress()
	notifyUrl := callbackAddr + "/api/bepusdt/notify"
	if strings.TrimSpace(setting.BepusdtNotifyUrl) != "" {
		notifyUrl = strings.TrimSpace(setting.BepusdtNotifyUrl)
	}
	returnUrl := paymentReturnPath("/console/topup?show_history=true")
	if strings.TrimSpace(setting.BepusdtReturnUrl) != "" {
		returnUrl = strings.TrimSpace(setting.BepusdtReturnUrl)
	}
	returnUrl = service.BepusdtSuccessRedirect(returnUrl)

	amountCents := service.BepusdtMoneyToCents(payMoney)
	amountJSON := service.BepusdtAmountJSON(amountCents)

	data := map[string]any{
		"order_id":     tradeNo,
		"amount":       amountJSON,
		"notify_url":   notifyUrl,
		"redirect_url": returnUrl,
		"fiat":         getBepusdtFiat(),
	}

	path := "/api/v1/order/create-order"
	if tradeType := strings.TrimSpace(setting.BepusdtTradeType); tradeType != "" {
		path = "/api/v1/order/create-transaction"
		data["trade_type"] = tradeType
		data["no_rate"] = false
	}
	if currencies := strings.TrimSpace(setting.BepusdtCurrencies); currencies != "" {
		data["currencies"] = currencies
	}

	data["signature"] = service.BepusdtMD5Sign(data, setting.BepusdtApiKey, []string{"signature"})

	body, err := common.Marshal(data)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("BEPUSDT 序列化请求失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	endpoint := service.BepusdtWithPath(setting.BepusdtUrl, path)
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("BEPUSDT 构建请求失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := service.GetHttpClient().Do(httpReq)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("BEPUSDT 请求网关失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("BEPUSDT 读取网关响应失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	var createResp bepusdtCreateResp
	if err := common.Unmarshal(respBody, &createResp); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("BEPUSDT 解析网关响应失败 user_id=%d trade_no=%s body=%q error=%q", id, tradeNo, string(respBody), err.Error()))
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}

	if createResp.StatusCode != 200 || createResp.Data.PaymentURL == "" {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("BEPUSDT 网关业务失败 user_id=%d trade_no=%s status_code=%d message=%q body=%q", id, tradeNo, createResp.StatusCode, createResp.Message, string(respBody)))
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		msg := "拉起支付失败"
		if createResp.Message != "" {
			msg = createResp.Message
		}
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": msg})
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("BEPUSDT 充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%.2f gateway_trade_id=%s", id, tradeNo, req.Amount, payMoney, createResp.Data.TradeID))
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"payment_url": createResp.Data.PaymentURL,
			"order_id":    tradeNo,
			"trade_id":    createResp.Data.TradeID,
		},
	})
}

// BepusdtNotify handles BEPUSDT payment callbacks.
func BepusdtNotify(c *gin.Context) {
	if !isBepusdtWebhookEnabled() {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("BEPUSDT webhook 被拒绝 reason=webhook_disabled path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	params, err := parseBepusdtNotifyParams(c)
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("BEPUSDT webhook 参数解析失败 path=%q client_ip=%s error=%q", c.Request.RequestURI, c.ClientIP(), err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	logger.LogInfo(c.Request.Context(), fmt.Sprintf("BEPUSDT webhook 收到请求 path=%q client_ip=%s method=%s params=%q", c.Request.RequestURI, c.ClientIP(), c.Request.Method, common.GetJsonString(params)))

	if len(params) == 0 {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("BEPUSDT webhook 参数为空 path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	sign := params["signature"]
	signData := service.BepusdtMapToSignData(params)
	expected := service.BepusdtMD5Sign(signData, setting.BepusdtApiKey, []string{"signature"})
	if !strings.EqualFold(sign, expected) {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("BEPUSDT webhook 验签失败 path=%q client_ip=%s expected=%s got=%s", c.Request.RequestURI, c.ClientIP(), expected, sign))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	if params["status"] != "2" {
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("BEPUSDT webhook 忽略非成功状态 path=%q status=%s client_ip=%s", c.Request.RequestURI, params["status"], c.ClientIP()))
		// Still acknowledge so the gateway stops retrying non-success events that are validly signed.
		_, _ = c.Writer.Write([]byte("ok"))
		return
	}

	orderID := params["order_id"]
	if orderID == "" {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("BEPUSDT webhook 缺少 order_id path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	// Optional amount check against stored order money.
	if amountRaw, ok := params["amount"]; ok && amountRaw != "" {
		topUp := model.GetTopUpByTradeNo(orderID)
		if topUp != nil {
			expectedCents := service.BepusdtMoneyToCents(topUp.Money)
			paidCents, parseErr := service.BepusdtParseAmountToCents(amountRaw)
			if parseErr != nil {
				logger.LogWarn(c.Request.Context(), fmt.Sprintf("BEPUSDT webhook 金额解析失败 trade_no=%s amount=%q error=%q", orderID, amountRaw, parseErr.Error()))
				_, _ = c.Writer.Write([]byte("fail"))
				return
			}
			if !service.BepusdtAmountsMatch(expectedCents, paidCents) {
				logger.LogWarn(c.Request.Context(), fmt.Sprintf("BEPUSDT webhook 金额不匹配 trade_no=%s expected_cents=%d paid_cents=%d", orderID, expectedCents, paidCents))
				_, _ = c.Writer.Write([]byte("fail"))
				return
			}
		}
	}

	LockOrder(orderID)
	defer UnlockOrder(orderID)

	if err := model.RechargeBepusdt(orderID, c.ClientIP()); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("BEPUSDT 到账失败 trade_no=%s client_ip=%s error=%q", orderID, c.ClientIP(), err.Error()))
		// Order may already be success (idempotent) — still return ok when mismatch is not the cause.
		if err.Error() == "充值失败，请稍后重试" {
			// Distinguish mismatch vs other by re-checking provider on a read.
			topUp := model.GetTopUpByTradeNo(orderID)
			if topUp != nil && topUp.Status == common.TopUpStatusSuccess {
				_, _ = c.Writer.Write([]byte("ok"))
				return
			}
		}
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("BEPUSDT 充值成功 trade_no=%s client_ip=%s", orderID, c.ClientIP()))
	_, _ = c.Writer.Write([]byte("ok"))
}

func parseBepusdtNotifyParams(c *gin.Context) (map[string]string, error) {
	params := map[string]string{}

	// Prefer JSON body when Content-Type is JSON.
	ct := strings.ToLower(c.GetHeader("Content-Type"))
	if c.Request.Method == http.MethodPost && strings.Contains(ct, "application/json") {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return nil, err
		}
		if len(body) > 0 {
			var raw map[string]any
			if err := common.Unmarshal(body, &raw); err != nil {
				return nil, err
			}
			for k, v := range raw {
				if s, ok := service.BepusdtValueForSign(v); ok {
					params[k] = s
				} else if v != nil {
					params[k] = fmt.Sprint(v)
				}
			}
			return params, nil
		}
	}

	if c.Request.Method == http.MethodPost {
		if err := c.Request.ParseForm(); err != nil {
			return nil, err
		}
		params = lo.Reduce(lo.Keys(c.Request.PostForm), func(r map[string]string, t string, _ int) map[string]string {
			r[t] = c.Request.PostForm.Get(t)
			return r
		}, map[string]string{})
		if len(params) > 0 {
			return params, nil
		}
	}

	params = lo.Reduce(lo.Keys(c.Request.URL.Query()), func(r map[string]string, t string, _ int) map[string]string {
		r[t] = c.Request.URL.Query().Get(t)
		return r
	}, map[string]string{})
	return params, nil
}
