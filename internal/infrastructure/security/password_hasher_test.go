package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHasher() *ArgonHasher {
	timeCost := uint32(2)
	memoryCost := uint32(64 * 1024)
	threads := uint8(4)
	keyLength := uint32(32)
	return NewArgonHasher(timeCost, memoryCost, keyLength, threads)
}

func TestArgonHasher_Verify(t *testing.T) {
	t.Run("equal_passwords", func(t *testing.T) {
		hasher := newTestHasher()
		password := "VerySecret"
		passHash, err := hasher.Hash(password)

		require.NoError(t, err)
		assert.NotEqual(t, password, passHash)

		result, err := hasher.Verify(passHash, password)
		assert.NoError(t, err)
		assert.True(t, result)
	})
	t.Run("invalid_password_hash", func(t *testing.T) {
		hasher := newTestHasher()
		_, err := hasher.Verify("invalidHash", "InvalidPassword")
		assert.Error(t, err)
	})
	t.Run("different_passwords", func(t *testing.T) {
		hasher := newTestHasher()
		password := "VerySecret"
		passHash, err := hasher.Hash(password)

		require.NoError(t, err)
		assert.NotEqual(t, password, passHash)

		result, err := hasher.Verify(passHash, "InvalidPassword")
		assert.NoError(t, err)
		assert.False(t, result)
	})
}
