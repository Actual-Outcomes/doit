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

	"github.com/Actual-Outcomes/doit/internal/auth"
	"github.com/google/uuid"
)

const (
	sessionCookieName = "doit_session"
	sessionDuration   = 24 * time.Hour

	projectCookieName = "doit_project"
	projectCookieAge  = 30 * 24 * time.Hour
)

// Session holds the decoded session data from a verified cookie.
type Session struct {
	TenantID uuid.UUID
	Expiry   time.Time
	IsAdmin  bool
}

// signCookie creates an HMAC-SHA256 signed cookie value.
// Format: base64(tenantID|expiryUnix|isAdmin).base64(hmac)
func signCookie(tenantID uuid.UUID, isAdmin bool, signingKey string) string {
	expiry := time.Now().Add(sessionDuration).Unix()
	adminFlag := "0"
	if isAdmin {
		adminFlag = "1"
	}
	payload := fmt.Sprintf("%s|%d|%s", tenantID.String(), expiry, adminFlag)
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

	fields := strings.Split(string(payloadBytes), "|")
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid payload format")
	}

	tenantID, err := uuid.Parse(fields[0])
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID")
	}

	expiryUnix, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry")
	}

	expiry := time.Unix(expiryUnix, 0)
	if time.Now().After(expiry) {
		return nil, fmt.Errorf("session expired")
	}

	// Parse admin flag (backward compatible: default false)
	isAdmin := false
	if len(fields) >= 3 && fields[2] == "1" {
		isAdmin = true
	}

	return &Session{TenantID: tenantID, Expiry: expiry, IsAdmin: isAdmin}, nil
}

// setSessionCookie writes an authenticated session cookie.
func setSessionCookie(w http.ResponseWriter, tenantID uuid.UUID, isAdmin bool, signingKey string) {
	value := signCookie(tenantID, isAdmin, signingKey)
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

// setProjectCookie writes the selected project ID cookie.
func setProjectCookie(w http.ResponseWriter, projectID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     projectCookieName,
		Value:    projectID,
		Path:     "/ui/",
		MaxAge:   int(projectCookieAge.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearProjectCookie expires the project cookie.
func clearProjectCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     projectCookieName,
		Value:    "",
		Path:     "/ui/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// getProjectCookie returns the project ID from the cookie, if set.
func getProjectCookie(r *http.Request) string {
	cookie, err := r.Cookie(projectCookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	return cookie.Value
}

// SessionMiddleware validates session cookies and redirects to login if invalid.
// It also injects project filtering into the context if a project cookie is set.
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

			// Inject project filter if project cookie is set
			if projectID := getProjectCookie(r); projectID != "" {
				ctx := auth.WithAllowedProjects(r.Context(), []string{projectID})
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdminSessionMiddleware validates that the session belongs to an admin user.
// Must be used after SessionMiddleware (session is already verified).
func AdminSessionMiddleware(signingKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				http.Redirect(w, r, "/ui/login", http.StatusFound)
				return
			}

			sess, err := verifyCookie(cookie.Value, signingKey)
			if err != nil {
				clearSessionCookie(w)
				http.Redirect(w, r, "/ui/login", http.StatusFound)
				return
			}

			if !sess.IsAdmin {
				http.Error(w, "Forbidden — admin access required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
