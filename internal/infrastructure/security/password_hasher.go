package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const saltSize = 16

type ArgonHasher struct {
	TimeCost   uint32
	MemoryCost uint32
	Threads    uint8
	KeyLength  uint32
}

func NewArgonHasher(timeCost, memoryCost, keyLength uint32, threads uint8) *ArgonHasher {
	return &ArgonHasher{
		TimeCost:   timeCost,
		MemoryCost: memoryCost,
		Threads:    threads,
		KeyLength:  keyLength,
	}
}

type ParsedHash struct {
	ArgonHasher
	HashRaw []byte
	Salt    []byte
}

func (h *ArgonHasher) Hash(password string) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", fmt.Errorf("password hashing failed: %w", err)
	}

	hashRaw := argon2.IDKey(
		[]byte(password),
		salt,
		h.TimeCost,
		h.MemoryCost,
		h.Threads,
		h.KeyLength,
	)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.MemoryCost,
		h.TimeCost,
		h.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hashRaw),
	)

	return encodedHash, nil
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("salt generation failed: %w", err)
	}
	return salt, nil
}

func (h *ArgonHasher) Verify(storedHash, providedPassword string) (bool, error) {
	config, err := parseArgon2Hash(storedHash)
	if err != nil {
		return false, fmt.Errorf("hash parsing failed: %w", err)
	}

	computedHash := argon2.IDKey(
		[]byte(providedPassword),
		config.Salt,
		config.TimeCost,
		config.MemoryCost,
		config.Threads,
		config.KeyLength,
	)

	match := subtle.ConstantTimeCompare(config.HashRaw, computedHash) == 1
	return match, nil
}

func parseArgon2Hash(encodedHash string) (*ParsedHash, error) {
	components := strings.Split(encodedHash, "$")
	if len(components) != 6 {
		return nil, errors.New("invalid hash format struct")
	}

	if !strings.HasPrefix(components[1], "argon2id") {
		return nil, errors.New("unsupported algorithm variant")
	}

	var version int
	if _, err := fmt.Sscanf(components[2], "v=%d", &version); err != nil {
		return nil, fmt.Errorf("error scan version: %w", err)
	}

	config := &ParsedHash{}
	if _, err := fmt.Sscanf(components[3], "m=%d,t=%d,p=%d",
		&config.MemoryCost, &config.TimeCost, &config.Threads); err != nil {
		return nil, fmt.Errorf("error scan config: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(components[4])
	if err != nil {
		return nil, fmt.Errorf("salt decoding failed: %w", err)
	}
	config.Salt = salt

	hash, err := base64.RawStdEncoding.DecodeString(components[5])
	if err != nil {
		return nil, fmt.Errorf("hash decoding failed %w", err)
	}
	config.HashRaw = hash
	config.KeyLength = uint32(len(hash))

	return config, nil
}
