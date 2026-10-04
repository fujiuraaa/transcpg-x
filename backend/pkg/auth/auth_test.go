package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNormalizePhone(t *testing.T) {
	ok := map[string]string{
		"0812-0000-0001":   "+6281200000001",
		"62 812 0000 0001": "+6281200000001",
		"+6281200000001":   "+6281200000001",
		"(0812) 0000.0001": "+6281200000001",
	}
	for in, want := range ok {
		if got, err := NormalizePhone(in); err != nil || got != want {
			t.Errorf("%q: dapat (%q, %v), ingin %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "812", "abc", "+0812", "08"} {
		if _, err := NormalizePhone(bad); !errors.Is(err, ErrInvalidPhone) {
			t.Errorf("%q harus ditolak", bad)
		}
	}
	if m := MaskPhone("+6281200000001"); m != "+62812****0001" {
		t.Errorf("mask: %s", m)
	}
}

func TestOTP(t *testing.T) {
	secret := []byte("rahasia")
	code, err := GenerateOTP()
	if err != nil || len(code) != OTPLength {
		t.Fatalf("kode %q, %v", code, err)
	}
	h := HashOTP(secret, 7, code)
	if !CheckOTP(secret, 7, code, h) {
		t.Fatal("kode benar harus lolos")
	}
	if CheckOTP(secret, 8, code, h) {
		t.Fatal("hash terikat ke pengguna — pengguna lain tidak boleh lolos")
	}
	if CheckOTP([]byte("lain"), 7, code, h) {
		t.Fatal("rahasia berbeda tidak boleh lolos")
	}
}

func TestGoogleVerifier(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("access_token") {
		case "ok":
			w.Write([]byte(`{"aud":"client-1","email":"Dokter@RS.test","email_verified":"true"}`))
		case "other-app":
			w.Write([]byte(`{"aud":"client-lain","email":"a@b.test","email_verified":"true"}`))
		case "unverified":
			w.Write([]byte(`{"aud":"client-1","email":"a@b.test","email_verified":"false"}`))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	g := NewGoogleVerifier("client-1")
	g.TokenInfoURL = srv.URL
	email, err := g.Verify(context.Background(), "ok")
	if err != nil || email != "dokter@rs.test" {
		t.Fatalf("dapat (%q, %v)", email, err)
	}
	for _, tok := range []string{"other-app", "unverified", "rusak", ""} {
		if _, err := g.Verify(context.Background(), tok); err == nil {
			t.Errorf("token %q harus ditolak", tok)
		}
	}
	if _, err := NewGoogleVerifier("").Verify(context.Background(), "ok"); err == nil {
		t.Error("tanpa client ID harus ditolak")
	}
}

func TestTokenVersion(t *testing.T) {
	secret := []byte("rahasia-uji-0123456789abcdef0123456789")
	tok, _, err := IssueToken(secret, 42, 3, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	id, ver, err := ParseToken(secret, tok)
	if err != nil || id != 42 || ver != 3 {
		t.Fatalf("parse: id=%d ver=%d err=%v", id, ver, err)
	}
}

func TestMaskPhoneShort(t *testing.T) {
	// Nomor E.164 terpendek (9 karakter) dulu membuat panic.
	if got := MaskPhone("+62123456"); got == "" {
		t.Fatal("hasil kosong")
	}
}

func TestPasswordTooLong(t *testing.T) {
	if _, err := HashPassword(strings.Repeat("a", 73)); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("ingin ErrPasswordTooLong, dapat %v", err)
	}
}

func TestDeriveKeyDiffers(t *testing.T) {
	m := []byte("induk")
	if string(DeriveKey(m, "otp")) == string(DeriveKey(m, "storage-local")) {
		t.Fatal("kunci turunan harus berbeda per keperluan")
	}
}
