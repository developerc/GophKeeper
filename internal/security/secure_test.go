package security

import (
	"testing"

	"github.com/developerc/GophKeeper/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecure(t *testing.T) {
	settings, err := config.NewServerSettings()
	require.NoError(t, err)
	var cipherText string = "cipher_text"
	var encriptedBytes []byte
	cipherManager, err := NewCipherManager(settings.Key)
	t.Run("#1_EncriptTest", func(t *testing.T) {
		encriptedBytes, err = cipherManager.Encrypt([]byte(cipherText))
		require.NoError(t, err)
	})

	t.Run("#2_DecriptTest", func(t *testing.T) {
		decriptedBytes, err := cipherManager.Decrypt(encriptedBytes)
		require.NoError(t, err)
		assert.Equal(t, decriptedBytes, []byte(cipherText), "cipher text and decripted should be equal")
	})

	t.Run("#3_GenerateJwtTest", func(t *testing.T) {
		jwtManager, err := NewJWTManager(settings.Key, settings.TokenDuration)
		require.NoError(t, err)
		_, err = jwtManager.GenerateJWT("UserID", "UserLogin")
		require.NoError(t, err)
	})
}
