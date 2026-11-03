package epusdtadapter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/modules/payment/infrastructure/gateway/epusdt"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"

	"github.com/shopspring/decimal"
)

func TestEpusdtAdapter_Type(t *testing.T) {
	a := NewEpusdtAdapter()
	want := constants.PaymentProviderEpusdt + ":"
	if got := a.Type(); got != want {
		t.Fatalf("Type() = %q, want %q", got, want)
	}
}

func TestEpusdtAdapter_ValidateConfig_EmptyRejected(t *testing.T) {
	a := NewEpusdtAdapter()
	err := a.ValidateConfig(jsonmap.JSON{}, "")
	if err == nil {
		t.Fatalf("expected error from empty config")
	}
	if !errors.Is(err, paymentcontract.ErrGatewayConfigInvalid) {
		t.Fatalf("expected wrapped paymentcontract.ErrGatewayConfigInvalid, got %v", err)
	}
}

func TestEpusdtAdapter_CreatePayment_ConfigInvalidMapped(t *testing.T) {
	a := NewEpusdtAdapter()
	_, err := a.CreatePayment(context.Background(), jsonmap.JSON{}, paymentcontract.GatewayCreateInput{
		OrderNo:  "ORDER_1",
		Currency: "USDT",
	})
	if err == nil {
		t.Fatalf("expected error from empty config")
	}
	if !errors.Is(err, paymentcontract.ErrGatewayConfigInvalid) {
		t.Fatalf("expected wrapped paymentcontract.ErrGatewayConfigInvalid, got %v", err)
	}
}

func TestEpusdtDisplayChannelType(t *testing.T) {
	tests := []struct {
		name string
		cfg  *epusdt.Config
		want string
	}{
		{name: "nil", cfg: nil, want: ""},
		{
			name: "cashier",
			cfg:  &epusdt.Config{OrderMode: constants.PaymentEpusdtOrderModeCashier},
			want: "",
		},
		{
			name: "tron usdt",
			cfg:  &epusdt.Config{Token: " USDT ", Network: " TRON "},
			want: "usdt.tron",
		},
		{
			name: "tron trx",
			cfg:  &epusdt.Config{Token: " trx ", Network: " tron "},
			want: "trx.tron",
		},
		{
			name: "fallback payment type",
			cfg:  &epusdt.Config{Token: "usdc", Network: "ethereum", PaymentType: "ethereum-usdc"},
			want: "usdc.ethereum",
		},
		{
			name: "unknown without payment type",
			cfg:  &epusdt.Config{Token: "usdc", Network: "ethereum"},
			want: "usdc.ethereum",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := epusdtDisplayChannelType(tc.cfg); got != tc.want {
				t.Fatalf("epusdtDisplayChannelType() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEpusdtAdapter_MapEpusdtError(t *testing.T) {
	cases := []struct {
		name string
		in   error
		want error
	}{
		{"config", epusdt.ErrConfigInvalid, paymentcontract.ErrGatewayConfigInvalid},
		{"request", epusdt.ErrRequestFailed, paymentcontract.ErrGatewayRequestFailed},
		{"response", epusdt.ErrResponseInvalid, paymentcontract.ErrGatewayResponseInvalid},
		{"signature", epusdt.ErrSignatureInvalid, paymentcontract.ErrGatewaySignatureInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapEpusdtError(tc.in)
			if !errors.Is(got, tc.want) {
				t.Fatalf("mapEpusdtError(%v) errors.Is %v = false, want true", tc.in, tc.want)
			}
		})
	}
}

func createEpusdtTestPayment(t *testing.T, gatewayResponse string, orderMode string, mode string) *paymentcontract.GatewayCreateResult {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(gatewayResponse))
	}))
	t.Cleanup(server.Close)

	result, err := NewEpusdtAdapter().CreatePayment(context.Background(), jsonmap.JSON{
		"gateway_url": server.URL,
		"pid":         "1000",
		"secret_key":  "sk-test",
		"order_mode":  orderMode,
		"token":       "USDT",
		"network":     "TRON",
		"notify_url":  "https://example.com/notify",
		"return_url":  "https://example.com/return",
	}, paymentcontract.GatewayCreateInput{
		OrderNo: "ORDER_1",
		Subject: "test",
		Amount:  money.FromDecimal(decimal.RequireFromString("28.88")),
		Extra:   jsonmap.JSON{"interaction_mode": mode},
	})
	if err != nil {
		t.Fatalf("CreatePayment() error = %v", err)
	}
	return result
}

func TestEpusdtAdapter_CreatePayment_QRModeUsesReceiveAddress(t *testing.T) {
	result := createEpusdtTestPayment(t,
		`{"status_code":200,"data":{"trade_id":"T1","actual_amount":"4.25","receive_address":" TAddr "}}`,
		constants.PaymentEpusdtOrderModeTransaction, constants.PaymentInteractionQR)

	if result.QRCodeURL != "TAddr" {
		t.Fatalf("QRCodeURL = %q, want the receive address", result.QRCodeURL)
	}
	if result.RedirectURL != "" {
		t.Fatalf("RedirectURL = %q, want empty in qr mode", result.RedirectURL)
	}
	data, _ := result.Payload["data"].(map[string]interface{})
	if data["chain"] != "tron" || data["token_id"] != "tron-usdt" {
		t.Fatalf("payload chain/token_id = %v / %v", data["chain"], data["token_id"])
	}
	if data["actual_amount"] != "4.25" {
		t.Fatalf("gateway fields must be preserved, got actual_amount = %v", data["actual_amount"])
	}
}

func TestEpusdtAdapter_CreatePayment_QRModeWithoutAddressKeepsPayLink(t *testing.T) {
	result := createEpusdtTestPayment(t,
		`{"status_code":200,"data":{"trade_id":"T1","actual_amount":"4.25"}}`,
		constants.PaymentEpusdtOrderModeTransaction, constants.PaymentInteractionQR)

	if result.QRCodeURL == "" || result.QRCodeURL != result.RedirectURL {
		t.Fatalf("expected pay link fallback, got qr=%q redirect=%q", result.QRCodeURL, result.RedirectURL)
	}
}

func TestEpusdtAdapter_CreatePayment_RedirectModeKeepsPayLink(t *testing.T) {
	result := createEpusdtTestPayment(t,
		`{"status_code":200,"data":{"trade_id":"T1","receive_address":"TAddr"}}`,
		constants.PaymentEpusdtOrderModeTransaction, constants.PaymentInteractionRedirect)

	if result.RedirectURL == "" || result.QRCodeURL != result.RedirectURL {
		t.Fatalf("redirect mode must keep pay link, got qr=%q redirect=%q", result.QRCodeURL, result.RedirectURL)
	}
}

func TestEpusdtAdapter_CreatePayment_CashierModeOmitsAssetLabels(t *testing.T) {
	result := createEpusdtTestPayment(t,
		`{"status_code":200,"data":{"trade_id":"T1"}}`,
		constants.PaymentEpusdtOrderModeCashier, constants.PaymentInteractionRedirect)

	data, _ := result.Payload["data"].(map[string]interface{})
	if _, ok := data["chain"]; ok {
		t.Fatalf("cashier mode must not claim a chain: %v", data)
	}
	if _, ok := data["token_id"]; ok {
		t.Fatalf("cashier mode must not claim a token: %v", data)
	}
}
