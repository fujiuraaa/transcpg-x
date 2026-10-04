// Package httpx berisi helper HTTP: respons JSON, galat terstruktur, dan
// penerjemahan galat database menjadi status HTTP.
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Error adalah galat yang aman ditampilkan ke pengguna.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *Error) Error() string { return e.Message }

func NewError(status int, code, msg string) *Error {
	return &Error{Status: status, Code: code, Message: msg}
}

func BadRequest(msg string) *Error   { return NewError(http.StatusBadRequest, "BAD_REQUEST", msg) }
func Unauthorized(msg string) *Error { return NewError(http.StatusUnauthorized, "UNAUTHORIZED", msg) }
func Forbidden(msg string) *Error    { return NewError(http.StatusForbidden, "FORBIDDEN", msg) }
func NotFound(msg string) *Error     { return NewError(http.StatusNotFound, "NOT_FOUND", msg) }
func Conflict(msg string) *Error     { return NewError(http.StatusConflict, "CONFLICT", msg) }
func Unprocessable(msg string) *Error {
	return NewError(http.StatusUnprocessableEntity, "VALIDATION", msg)
}
func NotImplemented(what string) *Error {
	return NewError(http.StatusNotImplemented, "NOT_IMPLEMENTED", what+" belum diimplementasikan")
}

// HandlerFunc adalah handler yang mengembalikan galat alih-alih menulisnya.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Handle membungkus HandlerFunc menjadi http.HandlerFunc.
func Handle(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			WriteError(w, r, err)
		}
	}
}

func JSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return nil
	}
	return json.NewEncoder(w).Encode(v)
}

func OK(w http.ResponseWriter, v any) error { return JSON(w, http.StatusOK, v) }

func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// Decode membaca body JSON (maks. 1 MB) dan menolak field yang tidak dikenal.
func Decode(r *http.Request, v any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return BadRequest("body JSON tidak valid: " + err.Error())
	}
	return nil
}

// PathInt64 membaca parameter path numerik, mis. {id}.
func PathInt64(r *http.Request, name string) (int64, error) {
	v, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil {
		return 0, BadRequest("parameter " + name + " tidak valid")
	}
	return v, nil
}

// QueryInt membaca query string numerik dengan nilai bawaan.
func QueryInt(r *http.Request, name string, def int) int {
	if v, err := strconv.Atoi(r.URL.Query().Get(name)); err == nil {
		return v
	}
	return def
}

// WriteError menulis galat sebagai JSON. Galat database yang dikenali
// diterjemahkan; galat lain dicatat dan dikembalikan sebagai 500.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var he *Error
	if errors.As(err, &he) {
		_ = JSON(w, he.Status, he)
		return
	}
	if errors.Is(err, pgx.ErrNoRows) {
		_ = JSON(w, http.StatusNotFound, NotFound("data tidak ditemukan"))
		return
	}
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		switch pe.Code {
		case "23505":
			_ = JSON(w, http.StatusConflict, &Error{Code: "DUPLIKAT", Message: "data yang sama sudah ada", Details: pe.ConstraintName})
			return
		case "23514", "23503", "23502", "P0001":
			e := &Error{Code: "VALIDATION", Message: pe.Message}
			if pe.ConstraintName != "" {
				e.Details = pe.ConstraintName
			}
			_ = JSON(w, http.StatusUnprocessableEntity, e)
			return
		}
	}
	slog.Error("galat tak tertangani", "method", r.Method, "path", r.URL.Path, "err", err)
	_ = JSON(w, http.StatusInternalServerError, NewError(500, "INTERNAL", "terjadi kesalahan pada server"))
}
