package tokens

import (
	"context"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
	"time"
)

var jwtSecret = []byte("your_jwt_secret")

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func CreateToken(userID, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	return t.SignedString(jwtSecret)
}

type contextKey string

const claimsKey = contextKey("claims")

func AuthMiddleware(h http.HandlerFunc, roles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get("Authorization")
		if hdr == "" {
			http.Error(w, `{"message":"missing auth"}`, http.StatusUnauthorized)
			return
		}
		tok := strings.TrimPrefix(hdr, "Bearer ")
		parsed, err := jwt.ParseWithClaims(tok, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !parsed.Valid {
			http.Error(w, `{"message":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		c := parsed.Claims.(*Claims)
		if len(roles) > 0 {
			ok := false
			for _, role := range roles {
				if c.Role == role {
					ok = true
				}
			}
			if !ok {
				http.Error(w, `{"message":"forbidden"}`, http.StatusForbidden)
				return
			}
		}
		r = r.WithContext(context.WithValue(r.Context(), claimsKey, c))
		h(w, r)
	}
}
