package service

import (
	"crypto/subtle"
	"fmt"
)

func fmtSscanf(s, format string, a ...any) (int, error) { return fmt.Sscanf(s, format, a...) }
func fmtError(s string) error                           { return fmt.Errorf("%s", s) }
func subtleCompare(a, b []byte) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare(a, b) == 1
}
