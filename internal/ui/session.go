package ui

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	sessionCookieName = "doit_session"
	sessionDuration   = 24 * time.Hour
)

// Session holds the decoded session data from a verified cookie.
type Session struct {
	TenantID uuid.UUID
	Expiry   time.Time
}

// signCookie creates an HMAC-SHA256 signed cookie value.
// Format: base64(tenantID|expiryUnix).base64(hmac)
func signCookie(tenantID uuid.UUID, signingKey string) string {
	expiry := time.Now().Add(sessionDuration).Unix()
	payload := fmt.Sprintf("%s|%d", tenantID.String(), expiry)
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))

	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(encodedPayload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return encodedPayload + "." + sig
}

// verifyCookie verifies the HMAC signature and checks expiry.
func verifyCookie(value, signingKey string) (*Session, error) {
	parts := strings.SplitN(value, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cookie format")
	}

	encodedPayload, encodedSig := parts[0], parts[1]

	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(encodedPayload))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(encodedSig), []byte(expectedSig)) {
		return nil, fmt.Errorf("invalid signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding")
	}

	payload := string(payloadBytes)
	sepIdx := strings.LastIndex(payload, "|")
	if sepIdx < 0 {
		return nil, fmt.Errorf("invalid payload format")
	}

	tenantIDStr := payload[:sepIdx]
	expiryStr := payload[sepIdx+1:]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID")
	}

	expiryUnix, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry")
	}

	expiry := time.Unix(expiryUnix, 0)
	if time.Now().After(expiry) {
		return nil, fmt.Errorf("session expired")
	}

	return &Session{TenantID: tenantID, Expiry: expiry}, nil
}

// setSessionCookie writes an authenticated session cookie.
func setSessionCookie(w http.ResponseWriter, tenantID uuid.UUID, signingKey string) {
	value := signCookie(tenantID, signingKey)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/ui/",
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSessionCookie expires the session cookie.
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/ui/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// SessionMiddleware validates session cookies and redirects to login if invalid.
func SessionMiddleware(signingKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				http.Redirect(w, r, "/ui/login", http.StatusFound)
				return
			}

			_, err = verifyCookie(cookie.Value, signingKey)
			if err != nil {
				clearSessionCookie(w)
				http.Redirect(w, r, "/ui/login", http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
