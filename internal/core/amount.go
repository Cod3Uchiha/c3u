package core

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	Coin     int64 = 100_000_000
	MaxMoney int64 = 21_000_000 * Coin
)

func ParseAmount(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "-") || strings.HasPrefix(s, "+") {
		return 0, fmt.Errorf("invalid C3U amount")
	}
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid C3U amount")
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return 0, fmt.Errorf("invalid C3U amount")
	}
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if len(frac) > 8 {
		return 0, fmt.Errorf("C3U supports at most 8 decimals")
	}
	for len(frac) < 8 {
		frac += "0"
	}
	fv := int64(0)
	if frac != "" {
		fv, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid C3U amount")
		}
	}
	if whole > MaxMoney/Coin {
		return 0, fmt.Errorf("amount exceeds maximum money")
	}
	v := whole*Coin + fv
	if v < 0 || v > MaxMoney {
		return 0, fmt.Errorf("amount exceeds maximum money")
	}
	return v, nil
}

func FormatAmount(v int64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	whole := v / Coin
	frac := v % Coin
	s := fmt.Sprintf("%d.%08d", whole, frac)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	if neg {
		s = "-" + s
	}
	return s
}
