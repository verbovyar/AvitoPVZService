package tokens

import (
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateToken_And_ParseClaims(t *testing.T) {
	tokenStr, err := CreateToken("user123", "moderator")
	if err != nil {
		t.Fatalf("CreateToken error: %v", err)
	}

	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		t.Fatal("Token is invalid or claims have wrong type")
	}
	if claims.UserID != "user123" || claims.Role != "moderator" {
		t.Errorf("Unexpected claims: got %v", claims)
	}
}

func TestAuthMiddleware_NoAuthHeader(t *testing.T) {
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when no auth header")
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	h(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid token")
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer bad.token.value")
	h(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_ForbiddenRole(t *testing.T) {
	tokenStr, _ := CreateToken("u1", "employee")
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for forbidden role")
	}, "moderator")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	h(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestAuthMiddleware_Success(t *testing.T) {
	tokenStr, _ := CreateToken("u42", "moderator")
	called := false
	h := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}, "moderator")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	h(rec, req)
	if !called {
		t.Error("handler was not called but should have been")
	}
}
