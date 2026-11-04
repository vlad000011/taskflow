package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
	"strings"
	"time"
)

func HashPassword(password string) string {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	hash := argon2.IDKey([]byte(password), salt, 2, 64*1024, 2, 32)
	return "argon2id$" + hex.EncodeToString(salt) + "$" + hex.EncodeToString(hash)
}
func CheckPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 3 || parts[0] != "argon2id" {
		return false
	}
	salt, e1 := hex.DecodeString(parts[1])
	expected, e2 := hex.DecodeString(parts[2])
	if e1 != nil || e2 != nil {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, 2, 64*1024, 2, uint32(len(expected)))
	return subtleCompare(got, expected)
}
func TokenHash(token string) string {
	s := sha256.Sum256([]byte(token))
	return hex.EncodeToString(s[:])
}
func MakeJWT(secret string, userID uint, role, kind string, expires time.Time) (string, error) {
	claims := jwt.MapClaims{"sub": userID, "role": role, "kind": kind, "exp": expires.Unix(), "iat": time.Now().Unix()}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
func ParseJWT(secret, token, kind string) (uint, string, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmtError("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return 0, "", fmtError("invalid token")
	}
	c, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || c["kind"] != kind {
		return 0, "", fmtError("invalid token kind")
	}
	sub, ok := c["sub"].(float64)
	if !ok || sub < 1 {
		return 0, "", fmtError("invalid subject")
	}
	role, _ := c["role"].(string)
	return uint(sub), role, nil
}
