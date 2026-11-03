package presenter

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

// CryptoWalletInfo 是二维码加密货币支付页需要直接展示的链上付款信息。
type CryptoWalletInfo struct {
	Address     string
	ChainAmount string
	Chain       string
	TokenID     string
}

// HasAny 返回是否存在任意可展示的链上付款字段。
func (info CryptoWalletInfo) HasAny() bool {
	return info.Address != "" || info.ChainAmount != "" || info.Chain != "" || info.TokenID != ""
}

// ExtractCryptoWalletInfo 从 Payment.ProviderPayload 中提取链上付款信息。
// 只在 interactionMode == "qr" 且 providerType 为加密货币网关时返回非空值；
// 其他情况返回空结构体，由调用方决定是否输出 omitempty 字段。
//
// 字段名按各 provider 原生响应：
//   - bepusdt:   data.token            (地址)  / data.actual_amount
//   - epusdt:    data.receive_address  (地址)  / data.actual_amount
//   - dujiaopay: pay_address           (地址)  / payable_amount / chain / token_id
//   - tokenpay:  暂不支持（包未解析地址），保留扩展位
func ExtractCryptoWalletInfo(providerType, interactionMode string, payload jsonmap.JSON) CryptoWalletInfo {
	if strings.ToLower(strings.TrimSpace(interactionMode)) != constants.PaymentInteractionQR {
		return CryptoWalletInfo{}
	}
	pt := strings.ToLower(strings.TrimSpace(providerType))
	switch pt {
	case constants.PaymentProviderBepusdt:
		return CryptoWalletInfo{
			Address:     readPayloadString(payload, "data", "token"),
			ChainAmount: readPayloadString(payload, "data", "actual_amount"),
			Chain:       readPayloadString(payload, "data", "chain"),
			TokenID:     readPayloadString(payload, "data", "token_id"),
		}
	case constants.PaymentProviderEpusdt:
		return CryptoWalletInfo{
			Address:     readPayloadString(payload, "data", "receive_address"),
			ChainAmount: readPayloadString(payload, "data", "actual_amount"),
			Chain:       readPayloadString(payload, "data", "chain"),
			TokenID:     readPayloadString(payload, "data", "token_id"),
		}
	case constants.PaymentProviderDujiaoPay:
		return CryptoWalletInfo{
			Address: firstPayloadString(
				readPayloadString(payload, "pay_address"),
				readPayloadString(payload, "data", "pay_address"),
			),
			ChainAmount: firstPayloadString(
				readPayloadString(payload, "payable_amount"),
				readPayloadString(payload, "data", "payable_amount"),
			),
			Chain: firstPayloadString(
				readPayloadString(payload, "chain"),
				readPayloadString(payload, "data", "chain"),
			),
			TokenID: firstPayloadString(
				readPayloadString(payload, "token_id"),
				readPayloadString(payload, "data", "token_id"),
			),
		}
	default:
		return CryptoWalletInfo{}
	}
}

// ExtractCryptoGatewayExpiresAt 返回网关停止收款监听的时间，取不到时返回 nil。
//
// 网关有独立于订单的收款时限（BEpusdt 未传 timeout 时默认 20 分钟），比订单更早到期时，
// 用户按订单倒计时转账可能无法自动到账，所以付款页要按较早的时间显示。
// data.expiration_time 在 BEpusdt 里是创建时的剩余秒数，也兼容绝对时间戳（秒或毫秒）。
// 这个值只用于展示，不参与支付状态判断。网关响应不可信，异常值一律视为取不到，
// 否则超出 JSON 时间范围的结果会让整个付款接口序列化失败。
func ExtractCryptoGatewayExpiresAt(providerType, interactionMode string, payload jsonmap.JSON, createdAt time.Time) *time.Time {
	if strings.ToLower(strings.TrimSpace(interactionMode)) != constants.PaymentInteractionQR {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(providerType)) {
	case constants.PaymentProviderBepusdt, constants.PaymentProviderEpusdt:
	default:
		return nil
	}
	const (
		unixSecondsFloor = 1_000_000_000
		unixMillisFloor  = 1_000_000_000_000
		unixMillisCap    = 1_000_000_000_000_000 // 超过这个值无论按哪种单位都没有意义
		maxJSONTimeYear  = 9999                  // time.Time 序列化为 JSON 的年份上限
	)
	value, err := strconv.ParseFloat(readPayloadString(payload, "data", "expiration_time"), 64)
	if err != nil || math.IsNaN(value) || value <= 0 || value >= unixMillisCap {
		return nil
	}
	var expiresAt time.Time
	switch {
	case value >= unixMillisFloor:
		expiresAt = time.UnixMilli(int64(value))
	case value >= unixSecondsFloor:
		expiresAt = time.Unix(int64(value), 0)
	case createdAt.IsZero():
		return nil
	default:
		expiresAt = createdAt.Add(time.Duration(value) * time.Second)
	}
	if expiresAt.Year() > maxJSONTimeYear {
		return nil
	}
	return &expiresAt
}

func firstPayloadString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// readPayloadString 沿 keys 路径在 payload 内取字符串值，支持 string / json.Number / 数值（fmt 转换）。
func readPayloadString(payload jsonmap.JSON, keys ...string) string {
	if payload == nil || len(keys) == 0 {
		return ""
	}
	var cur any = map[string]any(payload)
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			if mj, ok2 := cur.(jsonmap.JSON); ok2 {
				m = map[string]any(mj)
			} else {
				return ""
			}
		}
		v, ok := m[k]
		if !ok {
			return ""
		}
		cur = v
	}
	switch v := cur.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.8f", v), "0"), ".")
	case int, int32, int64:
		return strings.TrimSpace(fmt.Sprintf("%d", v))
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
