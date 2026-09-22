// Package argon2 implements the ports.PasswordHasher port with argon2id,
// tuned to the second recommended option of RFC 9106 (64 MiB, t=3) which
// fits the production VPS where Postgres, Mongo and MinIO coexist.
package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memoryKiB   = 64 * 1024
	iterations  = 3
	parallelism = 2
	saltLength  = 16
	keyLength   = 32
)

// Hasher hashes and verifies passwords using argon2id.
type Hasher struct{}

// NewHasher builds the argon2id hasher.
func NewHasher() *Hasher {
	return &Hasher{}
}

// Hash derives an argon2id key and encodes it as a PHC string
// ($argon2id$v=19$m=...,t=...,p=...$salt$hash), so the parameters travel
// with the hash and Verify needs no stored configuration.
func (Hasher) Hash(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, iterations, memoryKiB, parallelism, keyLength)

	var b strings.Builder
	b.WriteString("$argon2id$v=19$m=")
	b.WriteString(strconv.Itoa(memoryKiB))
	b.WriteString(",t=")
	b.WriteString(strconv.Itoa(iterations))
	b.WriteString(",p=")
	b.WriteString(strconv.Itoa(parallelism))
	b.WriteString("$")
	b.WriteString(base64.RawStdEncoding.EncodeToString(salt))
	b.WriteString("$")
	b.WriteString(base64.RawStdEncoding.EncodeToString(key))
	return b.String(), nil
}

// Verify re-derives the key using the parameters embedded in the PHC string
// and compares both keys in constant time.
func (Hasher) Verify(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, errors.New("invalid argon2id hash format")
	}

	var memory, time uint32
	var parallel uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &parallel); err != nil {
		return false, fmt.Errorf("parsing argon2id parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decoding salt: %w", err)
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decoding hash: %w", err)
	}

	// Derive with our fixed key length: a stored hash of any other length
	// simply fails the constant-time comparison below.
	got := argon2.IDKey([]byte(password), salt, time, memory, parallel, keyLength)
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
