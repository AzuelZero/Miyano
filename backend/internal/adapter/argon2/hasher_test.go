package argon2

import (
	"strings"
	"testing"
)

func TestHashVerifyRoundTrip(t *testing.T) {
	h := NewHasher()

	hash, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	ok, err := h.Verify("correct horse battery staple", hash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !ok {
		t.Error("Verify() = false for the same password, want true")
	}
}

func TestVerifyRejectsWrongPassword(t *testing.T) {
	h := NewHasher()

	hash, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	ok, err := h.Verify("Tr0ub4dor&3", hash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if ok {
		t.Error("Verify() = true for a different password, want false")
	}
}

func TestHashProducesPHCFormat(t *testing.T) {
	h := NewHasher()

	hash, err := h.Hash("some password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$v=19$m=65536,t=3,p=2$") {
		t.Errorf("hash = %q, want PHC prefix $argon2id$v=19$m=65536,t=3,p=2$", hash)
	}
	if strings.Contains(hash, "some password") {
		t.Error("hash leaks the plaintext password")
	}
}

func TestHashIsSalted(t *testing.T) {
	h := NewHasher()

	first, err := h.Hash("same password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	second, err := h.Hash("same password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if first == second {
		t.Error("two hashes of the same password are equal, want different salts")
	}
}

func TestVerifyRejectsMalformedHash(t *testing.T) {
	h := NewHasher()

	ok, err := h.Verify("password", "not-a-phc-string")
	if err == nil {
		t.Error("Verify() with malformed hash returned nil error, want error")
	}
	if ok {
		t.Error("Verify() = true for malformed hash, want false")
	}
}
