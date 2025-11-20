package wireguard

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeyPair(t *testing.T) {
	keyPair, err := GenerateKeyPair()

	require.NoError(t, err)
	require.NotNil(t, keyPair)

	// Check key lengths (44 characters in base64 = 32 bytes)
	assert.Equal(t, 44, len(keyPair.PrivateKey))
	assert.Equal(t, 44, len(keyPair.PublicKey))

	// Keys should be different
	assert.NotEqual(t, keyPair.PrivateKey, keyPair.PublicKey)

	// Should be valid base64
	_, err = base64.StdEncoding.DecodeString(keyPair.PrivateKey)
	assert.NoError(t, err)

	_, err = base64.StdEncoding.DecodeString(keyPair.PublicKey)
	assert.NoError(t, err)
}

func TestGenerateKeyPair_Uniqueness(t *testing.T) {
	keyPair1, err := GenerateKeyPair()
	require.NoError(t, err)

	keyPair2, err := GenerateKeyPair()
	require.NoError(t, err)

	// Each generation should produce unique keys
	assert.NotEqual(t, keyPair1.PrivateKey, keyPair2.PrivateKey)
	assert.NotEqual(t, keyPair1.PublicKey, keyPair2.PublicKey)
}

func TestPublicKeyFromPrivate(t *testing.T) {
	keyPair, err := GenerateKeyPair()
	require.NoError(t, err)

	// Derive public key from private key
	derivedPublicKey, err := PublicKeyFromPrivate(keyPair.PrivateKey)

	require.NoError(t, err)
	assert.Equal(t, keyPair.PublicKey, derivedPublicKey)
}

func TestPublicKeyFromPrivate_InvalidPrivateKey(t *testing.T) {
	tests := []struct {
		name       string
		privateKey string
		wantErr    string
	}{
		{
			name:       "Invalid base64",
			privateKey: "not-valid-base64!!!",
			wantErr:    "invalid private key encoding",
		},
		{
			name:       "Wrong length",
			privateKey: base64.StdEncoding.EncodeToString([]byte("short")),
			wantErr:    "invalid private key length",
		},
		{
			name:       "Empty string",
			privateKey: "",
			wantErr:    "invalid private key length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PublicKeyFromPrivate(tt.privateKey)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestValidatePrivateKey(t *testing.T) {
	t.Run("Valid private key", func(t *testing.T) {
		keyPair, err := GenerateKeyPair()
		require.NoError(t, err)

		err = ValidatePrivateKey(keyPair.PrivateKey)
		assert.NoError(t, err)
	})

	t.Run("Invalid base64", func(t *testing.T) {
		err := ValidatePrivateKey("invalid-base64!!!")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid private key encoding")
	})

	t.Run("Wrong length", func(t *testing.T) {
		shortKey := base64.StdEncoding.EncodeToString([]byte("short"))
		err := ValidatePrivateKey(shortKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid private key length")
	})

	t.Run("Empty string", func(t *testing.T) {
		err := ValidatePrivateKey("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid private key length")
	})
}

func TestValidatePublicKey(t *testing.T) {
	t.Run("Valid public key", func(t *testing.T) {
		keyPair, err := GenerateKeyPair()
		require.NoError(t, err)

		err = ValidatePublicKey(keyPair.PublicKey)
		assert.NoError(t, err)
	})

	t.Run("Invalid base64", func(t *testing.T) {
		err := ValidatePublicKey("invalid-base64!!!")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid public key encoding")
	})

	t.Run("Wrong length", func(t *testing.T) {
		shortKey := base64.StdEncoding.EncodeToString([]byte("short"))
		err := ValidatePublicKey(shortKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid public key length")
	})
}

func TestGeneratePresharedKey(t *testing.T) {
	psk, err := GeneratePresharedKey()

	require.NoError(t, err)
	assert.Equal(t, 44, len(psk))

	// Should be valid base64
	decoded, err := base64.StdEncoding.DecodeString(psk)
	require.NoError(t, err)
	assert.Equal(t, 32, len(decoded))
}

func TestGeneratePresharedKey_Uniqueness(t *testing.T) {
	psk1, err := GeneratePresharedKey()
	require.NoError(t, err)

	psk2, err := GeneratePresharedKey()
	require.NoError(t, err)

	// Each generation should produce unique keys
	assert.NotEqual(t, psk1, psk2)
}

func TestKeyPairStructure(t *testing.T) {
	keyPair := &KeyPair{
		PrivateKey: "test-private-key",
		PublicKey:  "test-public-key",
	}

	assert.Equal(t, "test-private-key", keyPair.PrivateKey)
	assert.Equal(t, "test-public-key", keyPair.PublicKey)
}

func TestCurve25519Compliance(t *testing.T) {
	// Generate multiple key pairs and verify they follow WireGuard spec
	for i := 0; i < 10; i++ {
		keyPair, err := GenerateKeyPair()
		require.NoError(t, err)

		// Decode private key
		privateKeyBytes, err := base64.StdEncoding.DecodeString(keyPair.PrivateKey)
		require.NoError(t, err)

		// Check clamping bits (WireGuard requirement)
		// privateKey[0] &= 248 means last 3 bits are 0
		assert.Equal(t, byte(0), privateKeyBytes[0]&7, "Private key should have last 3 bits of first byte set to 0")

		// privateKey[31] &= 127 means first bit is 0
		assert.Equal(t, byte(0), privateKeyBytes[31]&128, "Private key should have first bit of last byte set to 0")

		// privateKey[31] |= 64 means second bit is 1
		assert.Equal(t, byte(64), privateKeyBytes[31]&64, "Private key should have second bit of last byte set to 1")
	}
}
