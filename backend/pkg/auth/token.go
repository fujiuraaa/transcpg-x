// Package auth menangani login custom JWT: hash password, token, dan
// middleware yang memuat pengguna aktif ke context request.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const issuer = "transcpg-x"

type claims struct {
	jwt.RegisteredClaims
	// Version = app_users.token_version saat token dibuat. Dinaikkan saat kata
	// sandi diganti atau "keluar dari semua perangkat" → token lama ditolak.
	Version int `json:"ver"`
}

// IssueToken membuat JWT berisi ID pengguna dan versi sesi. Peran sengaja
// tidak disimpan di token: peran dibaca ulang dari database setiap request,
// sehingga perubahan peran atau penonaktifan akun berlaku seketika.
func IssueToken(secret []byte, userID int64, version int, ttl time.Duration) (string, time.Time, error) {
	exp := time.Now().Add(ttl)
	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
		Version: version,
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(secret)
	return tok, exp, err
}

// ParseToken memvalidasi JWT dan mengembalikan ID pengguna serta versi sesi.
func ParseToken(secret []byte, token string) (int64, int, error) {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c, func(*jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return 0, 0, err
	}
	id, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil {
		return 0, 0, errors.New("subject token tidak valid")
	}
	return id, c.Version, nil
}

// ErrPasswordTooLong: bcrypt hanya memakai 72 byte pertama.
var ErrPasswordTooLong = errors.New("kata sandi maksimal 72 karakter")

const MaxPasswordBytes = 72

func HashPassword(pw string) (string, error) {
	if len(pw) > MaxPasswordBytes {
		return "", ErrPasswordTooLong
	}
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

// dummyHash dipakai saat email tidak terdaftar agar waktu respons login sama
// dengan email terdaftar (tidak bisa dipakai menebak email).
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("transcpg-x-dummy-password"), bcrypt.DefaultCost)

// BurnPasswordCheck menjalankan bcrypt tanpa hasil, untuk menyamakan waktu respons.
func BurnPasswordCheck(pw string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(pw))
}

// DeriveKey menurunkan kunci terpisah per keperluan dari satu rahasia induk
// (JWT, OTP, token storage tidak berbagi kunci yang sama).
func DeriveKey(master []byte, purpose string) []byte {
	m := hmac.New(sha256.New, master)
	m.Write([]byte("transcpg-x/" + purpose))
	return m.Sum(nil)
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}
