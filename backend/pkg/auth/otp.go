package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

const (
	OTPLength         = 6
	OTPTTL            = 5 * time.Minute
	OTPMaxAttempts    = 5
	OTPResendInterval = 60 * time.Second
)

// GenerateOTP membuat kode angka acak sepanjang OTPLength (crypto/rand).
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", OTPLength, n.Int64()), nil
}

// HashOTP mengikat kode ke pengguna dengan HMAC-SHA256 berkunci rahasia
// server, sehingga hash di database tidak bisa ditebak dari 10⁶ kemungkinan.
func HashOTP(secret []byte, userID int64, code string) string {
	m := hmac.New(sha256.New, secret)
	fmt.Fprintf(m, "%d:%s", userID, code)
	return hex.EncodeToString(m.Sum(nil))
}

// CheckOTP membandingkan kode dengan hash secara waktu-konstan.
func CheckOTP(secret []byte, userID int64, code, hash string) bool {
	return hmac.Equal([]byte(HashOTP(secret, userID, code)), []byte(hash))
}
