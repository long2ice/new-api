package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// BepusdtValueForSign converts a JSON-like value into the string used for MD5 signing.
// Null and empty strings are omitted (ok=false).
func BepusdtValueForSign(v any) (string, bool) {
	switch value := v.(type) {
	case nil:
		return "", false
	case string:
		if value == "" {
			return "", false
		}
		return value, true
	case bool:
		if value {
			return "true", true
		}
		return "false", true
	case int:
		return strconv.Itoa(value), true
	case int64:
		return strconv.FormatInt(value, 10), true
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64), true
	default:
		s := fmt.Sprint(value)
		if s == "" || s == "<nil>" {
			return "", false
		}
		return s, true
	}
}

// BepusdtAmountJSON returns the amount value for the create-order body.
// Whole units become int64; fractional units become float64.
// Example: 900 cents → 9 as int64; 950 cents → 9.5 as float64.
func BepusdtAmountJSON(cents int64) any {
	if cents%100 == 0 {
		return cents / 100
	}
	whole := cents / 100
	frac := cents % 100
	if frac < 0 {
		frac = -frac
	}
	text := fmt.Sprintf("%d.%02d", whole, frac)
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return text
	}
	return f
}

// BepusdtMoneyToCents converts a major-unit money float to cents (rounded).
func BepusdtMoneyToCents(money float64) int64 {
	d := decimal.NewFromFloat(money).Mul(decimal.NewFromInt(100))
	return d.Round(0).IntPart()
}

// BepusdtParseAmountToCents parses a callback amount string into cents.
func BepusdtParseAmountToCents(raw string) (int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, fmt.Errorf("金额为空")
	}
	if whole, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return whole * 100, nil
	}
	d, err := decimal.NewFromString(trimmed)
	if err != nil {
		return 0, fmt.Errorf("无法解析金额: %s", trimmed)
	}
	if d.IsNegative() {
		return 0, fmt.Errorf("金额无效: %s", trimmed)
	}
	return d.Mul(decimal.NewFromInt(100)).Round(0).IntPart(), nil
}

// BepusdtAmountsMatch allows ±1 cent tolerance.
func BepusdtAmountsMatch(expectedCents, paidCents int64) bool {
	diff := expectedCents - paidCents
	if diff < 0 {
		diff = -diff
	}
	return diff <= 1
}

// BepusdtMD5Sign builds the MD5 signature for BEPUSDT create/notify payloads.
// skip keys (e.g. "signature") are excluded. Empty/null values are skipped.
func BepusdtMD5Sign(data map[string]any, apiKey string, skip []string) string {
	skipSet := make(map[string]struct{}, len(skip))
	for _, k := range skip {
		skipSet[k] = struct{}{}
	}

	type pair struct{ k, v string }
	pairs := make([]pair, 0, len(data))
	for k, v := range data {
		if _, ok := skipSet[k]; ok {
			continue
		}
		s, ok := BepusdtValueForSign(v)
		if !ok {
			continue
		}
		pairs = append(pairs, pair{k: k, v: s})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })

	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = p.k + "=" + p.v
	}
	content := strings.Join(parts, "&") + apiKey
	sum := md5.Sum([]byte(content))
	return hex.EncodeToString(sum[:])
}

// BepusdtMapToSignData converts a string map (callback params) into sign data.
// Values that parse as int64 become int64 so the sign string matches JSON numbers.
func BepusdtMapToSignData(data map[string]string) map[string]any {
	out := make(map[string]any, len(data))
	for k, v := range data {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			out[k] = n
			continue
		}
		out[k] = v
	}
	return out
}

// BepusdtSuccessRedirect appends trade_status=TRADE_SUCCESS to a return URL.
func BepusdtSuccessRedirect(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}
	sep := "?"
	if strings.Contains(rawURL, "?") {
		sep = "&"
	}
	return rawURL + sep + "trade_status=TRADE_SUCCESS"
}

// BepusdtWithPath joins base URL and path without double slashes.
func BepusdtWithPath(base, path string) string {
	return strings.TrimRight(base, "/") + path
}
