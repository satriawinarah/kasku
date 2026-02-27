package components

import (
	"fmt"
	"time"
)

// FormatMoney formats a float as a dot-separated integer string (Indonesian style).
// Example: 1500000 → "1.500.000"
func FormatMoney(v float64) string {
	if v < 0 {
		return "-" + FormatMoney(-v)
	}
	s := fmt.Sprintf("%.0f", v)
	out := []byte{}
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

// FormatMonth converts "2006-01" to "January 2006".
func FormatMonth(m string) string {
	t, err := time.Parse("2006-01", m)
	if err != nil {
		return m
	}
	return t.Format("January 2006")
}

// TodayISO returns today's date as "2006-01-02".
func TodayISO() string {
	return time.Now().Format("2006-01-02")
}
