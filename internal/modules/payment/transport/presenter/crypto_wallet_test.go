package presenter

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

func TestExtractUSDTWalletInfo(t *testing.T) {
	tests := []struct {
		name            string
		providerType    string
		interactionMode string
		payload         jsonmap.JSON
		wantAddress     string
		wantChainAmount string
	}{
		{
			name:            "bepusdt qr with token in data",
			providerType:    constants.PaymentProviderBepusdt,
			interactionMode: constants.PaymentInteractionQR,
			payload: jsonmap.JSON{
				"data": map[string]any{
					"token":         "TRX_ADDR_XYZ",
					"actual_amount": "13.45",
				},
			},
			wantAddress:     "TRX_ADDR_XYZ",
			wantChainAmount: "13.45",
		},
		{
			name:            "bepusdt redirect mode returns empty",
			providerType:    constants.PaymentProviderBepusdt,
			interactionMode: constants.PaymentInteractionRedirect,
			payload: jsonmap.JSON{
				"data": map[string]any{"token": "TRX_ADDR_XYZ", "actual_amount": "13.45"},
			},
			wantAddress:     "",
			wantChainAmount: "",
		},
		{
			name:            "epusdt qr without receive_address yields empty",
			providerType:    constants.PaymentProviderEpusdt,
			interactionMode: constants.PaymentInteractionQR,
			payload: jsonmap.JSON{
				"data": map[string]any{"trade_id": "T1", "actual_amount": "5.00"},
			},
			wantAddress:     "",
			wantChainAmount: "5.00",
		},
		{
			name:            "epusdt qr with receive_address",
			providerType:    constants.PaymentProviderEpusdt,
			interactionMode: constants.PaymentInteractionQR,
			payload: jsonmap.JSON{
				"data": map[string]any{"receive_address": "TGRC_ADDR", "actual_amount": "9.99"},
			},
			wantAddress:     "TGRC_ADDR",
			wantChainAmount: "9.99",
		},
		{
			name:            "non-usdt provider returns empty",
			providerType:    constants.PaymentProviderOfficial,
			interactionMode: constants.PaymentInteractionQR,
			payload:         jsonmap.JSON{"data": map[string]any{"token": "X"}},
			wantAddress:     "",
			wantChainAmount: "",
		},
		{
			name:            "nil payload safe",
			providerType:    constants.PaymentProviderBepusdt,
			interactionMode: constants.PaymentInteractionQR,
			payload:         nil,
			wantAddress:     "",
			wantChainAmount: "",
		},
		{
			name:            "bepusdt qr with data scalar instead of map",
			providerType:    constants.PaymentProviderBepusdt,
			interactionMode: constants.PaymentInteractionQR,
			payload:         jsonmap.JSON{"data": "oops"},
			wantAddress:     "",
			wantChainAmount: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info := ExtractCryptoWalletInfo(tc.providerType, tc.interactionMode, tc.payload)
			gotAddr, gotAmount := info.Address, info.ChainAmount
			if gotAddr != tc.wantAddress {
				t.Errorf("address: got %q want %q", gotAddr, tc.wantAddress)
			}
			if gotAmount != tc.wantChainAmount {
				t.Errorf("chain amount: got %q want %q", gotAmount, tc.wantChainAmount)
			}
		})
	}
}

func TestExtractCryptoWalletInfo_DujiaoPayQR(t *testing.T) {
	info := ExtractCryptoWalletInfo(constants.PaymentProviderDujiaoPay, constants.PaymentInteractionQR, jsonmap.JSON{
		"chain":          "tron",
		"token_id":       "tron-usdt",
		"pay_address":    "TAddr",
		"payable_amount": "10.0001",
	})

	if info.Address != "TAddr" {
		t.Fatalf("Address = %q", info.Address)
	}
	if info.ChainAmount != "10.0001" {
		t.Fatalf("ChainAmount = %q", info.ChainAmount)
	}
	if info.Chain != "tron" {
		t.Fatalf("Chain = %q", info.Chain)
	}
	if info.TokenID != "tron-usdt" {
		t.Fatalf("TokenID = %q", info.TokenID)
	}
}

func TestExtractCryptoWalletInfo_BepusdtQRIncludesChainLabels(t *testing.T) {
	info := ExtractCryptoWalletInfo(constants.PaymentProviderBepusdt, constants.PaymentInteractionQR, jsonmap.JSON{
		"data": map[string]any{
			"token":         "TAddr",
			"actual_amount": "4.25",
			"chain":         "tron",
			"token_id":      "tron-usdt",
		},
	})

	if info.Address != "TAddr" || info.ChainAmount != "4.25" || info.Chain != "tron" || info.TokenID != "tron-usdt" {
		t.Fatalf("unexpected info: %+v", info)
	}
}

func TestExtractCryptoWalletInfo_DujiaoPayWrappedPayload(t *testing.T) {
	info := ExtractCryptoWalletInfo(constants.PaymentProviderDujiaoPay, constants.PaymentInteractionQR, jsonmap.JSON{
		"data": map[string]any{
			"chain":          "base",
			"token_id":       "base-usdc",
			"pay_address":    "0xAddr",
			"payable_amount": "8.50",
		},
	})

	if info.Address != "0xAddr" || info.ChainAmount != "8.50" || info.Chain != "base" || info.TokenID != "base-usdc" {
		t.Fatalf("unexpected info: %+v", info)
	}
}

func TestExtractCryptoWalletInfo_EpusdtQRIncludesChainLabels(t *testing.T) {
	info := ExtractCryptoWalletInfo(constants.PaymentProviderEpusdt, constants.PaymentInteractionQR, jsonmap.JSON{
		"data": map[string]any{
			"receive_address": "TAddr",
			"actual_amount":   "4.25",
			"chain":           "tron",
			"token_id":        "tron-usdt",
		},
	})

	if info.Address != "TAddr" || info.ChainAmount != "4.25" || info.Chain != "tron" || info.TokenID != "tron-usdt" {
		t.Fatalf("unexpected info: %+v", info)
	}
}

func TestExtractCryptoGatewayExpiresAt(t *testing.T) {
	createdAt := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	payloadWith := func(value any) jsonmap.JSON {
		return jsonmap.JSON{"data": map[string]any{"expiration_time": value}}
	}
	absolute := createdAt.Add(30 * time.Minute)

	tests := []struct {
		name      string
		provider  string
		mode      string
		payload   jsonmap.JSON
		createdAt time.Time
		want      *time.Time
	}{
		{name: "relative seconds number", provider: constants.PaymentProviderBepusdt, mode: constants.PaymentInteractionQR, payload: payloadWith(float64(1199)), createdAt: createdAt, want: ptrTime(createdAt.Add(1199 * time.Second))},
		{name: "relative seconds string", provider: constants.PaymentProviderEpusdt, mode: constants.PaymentInteractionQR, payload: payloadWith("600"), createdAt: createdAt, want: ptrTime(createdAt.Add(600 * time.Second))},
		{name: "absolute unix seconds", provider: constants.PaymentProviderEpusdt, mode: constants.PaymentInteractionQR, payload: payloadWith(float64(absolute.Unix())), createdAt: createdAt, want: &absolute},
		{name: "absolute unix millis", provider: constants.PaymentProviderEpusdt, mode: constants.PaymentInteractionQR, payload: payloadWith(float64(absolute.UnixMilli())), createdAt: createdAt, want: &absolute},
		{name: "relative without created at", provider: constants.PaymentProviderBepusdt, mode: constants.PaymentInteractionQR, payload: payloadWith(float64(600))},
		{name: "missing field", provider: constants.PaymentProviderBepusdt, mode: constants.PaymentInteractionQR, payload: jsonmap.JSON{"data": map[string]any{}}, createdAt: createdAt},
		{name: "zero", provider: constants.PaymentProviderBepusdt, mode: constants.PaymentInteractionQR, payload: payloadWith(float64(0)), createdAt: createdAt},
		{name: "not numeric", provider: constants.PaymentProviderBepusdt, mode: constants.PaymentInteractionQR, payload: payloadWith("soon"), createdAt: createdAt},
		{name: "redirect mode ignored", provider: constants.PaymentProviderBepusdt, mode: constants.PaymentInteractionRedirect, payload: payloadWith(float64(600)), createdAt: createdAt},
		{name: "other provider ignored", provider: constants.PaymentProviderDujiaoPay, mode: constants.PaymentInteractionQR, payload: payloadWith(float64(600)), createdAt: createdAt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractCryptoGatewayExpiresAt(tt.provider, tt.mode, tt.payload, tt.createdAt)
			switch {
			case tt.want == nil && got != nil:
				t.Fatalf("got %v, want nil", got)
			case tt.want != nil && (got == nil || !got.Equal(*tt.want)):
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// 网关响应不可信：异常值不能让付款接口序列化失败，也不能把支付显示成已到期。
func TestExtractCryptoGatewayExpiresAt_RejectsAbnormalValues(t *testing.T) {
	createdAt := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	for _, value := range []any{"NaN", "Inf", "-Inf", "-1", "1e18", "1e30", float64(1e18), "999999999999", "1e15"} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			got := ExtractCryptoGatewayExpiresAt(constants.PaymentProviderBepusdt, constants.PaymentInteractionQR, jsonmap.JSON{
				"data": map[string]any{"expiration_time": value},
			}, createdAt)
			if got != nil {
				t.Fatalf("got %v, want nil", got)
			}
		})
	}
}

func TestExtractCryptoGatewayExpiresAt_ResultAlwaysSerializes(t *testing.T) {
	createdAt := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	for _, value := range []any{float64(1), "999999999", float64(1_000_000_000), "253402300799", float64(1_000_000_000_000), "999999999999999"} {
		got := ExtractCryptoGatewayExpiresAt(constants.PaymentProviderBepusdt, constants.PaymentInteractionQR, jsonmap.JSON{
			"data": map[string]any{"expiration_time": value},
		}, createdAt)
		if _, err := json.Marshal(LatestPaymentResp{GatewayExpiresAt: got}); err != nil {
			t.Fatalf("expiration_time %v: response must serialize, got %v", value, err)
		}
	}
}

func ptrTime(value time.Time) *time.Time { return &value }
