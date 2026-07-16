package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBepusdtAmountJSON(t *testing.T) {
	assert.Equal(t, int64(990), BepusdtAmountJSON(99000))
	assert.Equal(t, int64(9), BepusdtAmountJSON(900))
	assert.Equal(t, 9.5, BepusdtAmountJSON(950))
	assert.Equal(t, 0.01, BepusdtAmountJSON(1))
}

func TestBepusdtParseAmountToCents(t *testing.T) {
	cents, err := BepusdtParseAmountToCents("9")
	require.NoError(t, err)
	assert.Equal(t, int64(900), cents)

	cents, err = BepusdtParseAmountToCents("9.00")
	require.NoError(t, err)
	assert.Equal(t, int64(900), cents)

	cents, err = BepusdtParseAmountToCents("9.5")
	require.NoError(t, err)
	assert.Equal(t, int64(950), cents)

	cents, err = BepusdtParseAmountToCents("0.01")
	require.NoError(t, err)
	assert.Equal(t, int64(1), cents)
}

func TestBepusdtAmountsMatch(t *testing.T) {
	assert.True(t, BepusdtAmountsMatch(900, 900))
	assert.True(t, BepusdtAmountsMatch(900, 901))
	assert.False(t, BepusdtAmountsMatch(900, 902))
}

func TestBepusdtMD5SignStable(t *testing.T) {
	data := map[string]any{
		"order_id":     "o1",
		"amount":       int64(990),
		"notify_url":   "https://x/n",
		"redirect_url": "https://x/r?trade_status=TRADE_SUCCESS",
		"fiat":         "USD",
		"signature":    "should-be-skipped",
	}
	sign1 := BepusdtMD5Sign(data, "secret", []string{"signature"})
	sign2 := BepusdtMD5Sign(data, "secret", []string{"signature"})
	assert.Equal(t, sign1, sign2)
	assert.Len(t, sign1, 32)

	// Changing a field must change the signature.
	data["amount"] = int64(991)
	sign3 := BepusdtMD5Sign(data, "secret", []string{"signature"})
	assert.NotEqual(t, sign1, sign3)
}

func TestBepusdtMapToSignDataIntPromotion(t *testing.T) {
	data := map[string]string{
		"status":  "2",
		"amount":  "990",
		"order_id": "o1",
	}
	signData := BepusdtMapToSignData(data)
	assert.Equal(t, int64(2), signData["status"])
	assert.Equal(t, int64(990), signData["amount"])
	assert.Equal(t, "o1", signData["order_id"])
}

func TestBepusdtSuccessRedirect(t *testing.T) {
	assert.Equal(t, "https://x/r?trade_status=TRADE_SUCCESS", BepusdtSuccessRedirect("https://x/r"))
	assert.Equal(t, "https://x/r?a=1&trade_status=TRADE_SUCCESS", BepusdtSuccessRedirect("https://x/r?a=1"))
}
