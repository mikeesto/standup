package main

import (
	"fmt"
	"time"
)

func ordinal(n int) string {
	if n%100 >= 11 && n%100 <= 13 {
		return "th"
	}
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

func formatDateWithOrdinal(t time.Time) string {
	day := t.Day()
	return fmt.Sprintf("%s %d%s %s", t.Format("Monday"), day, ordinal(day), t.Format("January"))
}
