package auth

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidPhone = errors.New("nomor telepon tidak valid")
	e164            = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)
	phoneNoise      = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", ".", "")
)

// NormalizePhone mengubah nomor ke format E.164. Nomor Indonesia boleh ditulis
// 0812…, 62812…, atau +62812….
func NormalizePhone(s string) (string, error) {
	p := phoneNoise.Replace(strings.TrimSpace(s))
	switch {
	case strings.HasPrefix(p, "+"):
	case strings.HasPrefix(p, "62"):
		p = "+" + p
	case strings.HasPrefix(p, "0"):
		p = "+62" + p[1:]
	default:
		return "", ErrInvalidPhone
	}
	if !e164.MatchString(p) {
		return "", ErrInvalidPhone
	}
	return p, nil
}

// MaskPhone menyamarkan nomor untuk ditampilkan: +62812****0001.
func MaskPhone(p string) string {
	if len(p) < 8 {
		return p
	}
	return p[:6] + strings.Repeat("*", max(0, len(p)-10)) + p[len(p)-4:]
}
