package service

import (
	"strings"
	"testing"
	"time"
)

func TestHashPasswordAndCheckPassword(t *testing.T) {
	password := "correct horse battery staple"
	encoded := HashPassword(password)
	if encoded == "" || encoded == password {
		t.Fatal("HashPassword must return an encoded hash")
	}
	if !strings.HasPrefix(encoded, "argon2id$") {
		t.Fatalf("unexpected hash format: %q", encoded)
	}
	if !CheckPassword(encoded, password) {
		t.Fatal("CheckPassword rejected the correct password")
	}
	if CheckPassword(encoded, "incorrect password") {
		t.Fatal("CheckPassword accepted an incorrect password")
	}
	if encoded2 := HashPassword(password); encoded2 == encoded {
		t.Fatal("HashPassword should use a fresh random salt")
	}
}

func TestCheckPasswordRejectsMalformedHashes(t *testing.T) {
	cases := []string{
		"",
		"bcrypt$abc$def",
		"argon2id$not-hex$00",
		"argon2id$00$not-hex",
		"argon2id$00",
		"argon2id$$00",
	}
	for _, encoded := range cases {
		t.Run(encoded, func(t *testing.T) {
			if CheckPassword(encoded, "password") {
				t.Fatal("malformed hash must be rejected")
			}
		})
	}
}

func TestTokenHash(t *testing.T) {
	first := TokenHash("refresh-token")
	if len(first) != 64 {
		t.Fatalf("expected SHA-256 hex digest of 64 chars, got %d", len(first))
	}
	if first != TokenHash("refresh-token") {
		t.Fatal("TokenHash must be deterministic")
	}
	if first == TokenHash("another-token") {
		t.Fatal("different tokens must have different hashes")
	}
}

func TestMakeJWTAndParseJWT(t *testing.T) {
	secret := "test-secret-with-enough-entropy"
	expires := time.Now().Add(time.Hour)
	token, err := MakeJWT(secret, 42, "participant", "access", expires)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}
	id, role, err := ParseJWT(secret, token, "access")
	if err != nil {
		t.Fatalf("ParseJWT rejected a valid token: %v", err)
	}
	if id != 42 || role != "participant" {
		t.Fatalf("unexpected claims: id=%d role=%q", id, role)
	}
}

func TestParseJWTRejectsInvalidTokens(t *testing.T) {
	secret := "test-secret-with-enough-entropy"
	valid, err := MakeJWT(secret, 7, "lead", "access", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	expired, err := MakeJWT(secret, 7, "lead", "access", time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	wrongKind, err := MakeJWT(secret, 7, "lead", "refresh", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name   string
		secret string
		token  string
		kind   string
	}{
		{name: "malformed", secret: secret, token: "not-a-jwt", kind: "access"},
		{name: "wrong secret", secret: "different-secret", token: valid, kind: "access"},
		{name: "expired", secret: secret, token: expired, kind: "access"},
		{name: "wrong token kind", secret: secret, token: wrongKind, kind: "access"},
		{name: "empty token", secret: secret, token: "", kind: "access"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := ParseJWT(tc.secret, tc.token, tc.kind); err == nil {
				t.Fatal("expected token to be rejected")
			}
		})
	}
}

func TestSecurityHelpers(t *testing.T) {
	var parsed int
	if n, err := fmtSscanf("123", "%d", &parsed); err != nil || n != 1 || parsed != 123 {
		t.Fatalf("fmtSscanf result: n=%d parsed=%d err=%v", n, parsed, err)
	}
	if _, err := fmtSscanf("not-a-number", "%d", &parsed); err == nil {
		t.Fatal("fmtSscanf should report invalid input")
	}
	if fmtError("expected").Error() != "expected" {
		t.Fatal("fmtError returned unexpected message")
	}
	if !subtleCompare([]byte("same"), []byte("same")) {
		t.Fatal("subtleCompare should accept equal byte slices")
	}
	if subtleCompare([]byte("same"), []byte("diff")) {
		t.Fatal("subtleCompare should reject different byte slices")
	}
	if subtleCompare([]byte("short"), []byte("longer")) {
		t.Fatal("subtleCompare should reject different lengths")
	}
}
