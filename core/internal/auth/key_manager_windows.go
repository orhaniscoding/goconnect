//go:build windows

package auth

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// PrivateKey is a type alias for WireGuard private keys (32 bytes).
// On Windows, we handle keys as raw bytes to avoid the wgctrl dependency
// which requires Linux netlink APIs.
type PrivateKey [32]byte

// String returns the base64 representation of the key.
func (k PrivateKey) String() string {
	return encodeBase64(k[:])
}

// persistKey saves the private key to disk using DPAPI encryption.
func persistKey(path string, privKey PrivateKey) error {
	data := []byte(privKey.String())

	// Encrypt using DPAPI
	encrypted, err := encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt key with DPAPI: %w", err)
	}

	err = os.WriteFile(path, encrypted, 0600)
	if err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}
	return nil
}

// loadKey reads and decrypts the private key from disk.
func loadKey(path string) (PrivateKey, error) {
	encrypted, err := os.ReadFile(path)
	if err != nil {
		return PrivateKey{}, err
	}

	// Decrypt using DPAPI
	decrypted, err := decrypt(encrypted)
	if err != nil {
		return PrivateKey{}, fmt.Errorf("failed to decrypt key with DPAPI: %w", err)
	}

	privKey, err := parseKey(string(decrypted))
	if err != nil {
		return PrivateKey{}, fmt.Errorf("failed to parse key: %w", err)
	}

	return privKey, nil
}

// parseKey parses a base64-encoded WireGuard key into a PrivateKey.
func parseKey(s string) (PrivateKey, error) {
	b, err := decodeBase64(s)
	if err != nil {
		return PrivateKey{}, fmt.Errorf("failed to decode key: %w", err)
	}
	if len(b) != 32 {
		return PrivateKey{}, fmt.Errorf("invalid key length: got %d, want 32", len(b))
	}
	var key PrivateKey
	copy(key[:], b)
	return key, nil
}

// encodeBase64 encodes bytes to standard base64.
func encodeBase64(data []byte) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	result := make([]byte, ((len(data)+2)/3)*4)
	for i, j := 0, 0; i < len(data); i, j = i+3, j+4 {
		var val uint32
		switch len(data) - i {
		case 1:
			val = uint32(data[i]) << 16
			result[j] = base64Chars[val>>18&0x3F]
			result[j+1] = base64Chars[val>>12&0x3F]
			result[j+2] = '='
			result[j+3] = '='
		case 2:
			val = uint32(data[i])<<16 | uint32(data[i+1])<<8
			result[j] = base64Chars[val>>18&0x3F]
			result[j+1] = base64Chars[val>>12&0x3F]
			result[j+2] = base64Chars[val>>6&0x3F]
			result[j+3] = '='
		default:
			val = uint32(data[i])<<16 | uint32(data[i+1])<<8 | uint32(data[i+2])
			result[j] = base64Chars[val>>18&0x3F]
			result[j+1] = base64Chars[val>>12&0x3F]
			result[j+2] = base64Chars[val>>6&0x3F]
			result[j+3] = base64Chars[val&0x3F]
		}
	}
	return string(result)
}

// decodeBase64 decodes standard base64 to bytes.
func decodeBase64(s string) ([]byte, error) {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	// Build lookup table
	lookup := [256]int{}
	for i := range lookup {
		lookup[i] = -1
	}
	for i, c := range base64Chars {
		lookup[c] = i
	}

	// Remove padding and calculate output length
	for len(s) > 0 && s[len(s)-1] == '=' {
		s = s[:len(s)-1]
	}

	result := make([]byte, len(s)*6/8)
	bits := 0
	val := 0
	j := 0

	for _, c := range s {
		v := lookup[c]
		if v == -1 {
			return nil, fmt.Errorf("invalid base64 character: %c", c)
		}
		val = val<<6 | v
		bits += 6
		if bits >= 8 {
			bits -= 8
			result[j] = byte(val >> bits)
			j++
		}
	}

	return result[:j], nil
}

func encrypt(data []byte) ([]byte, error) {
	var in, out windows.DataBlob
	in.Data = &data[0]
	in.Size = uint32(len(data))

	err := windows.CryptProtectData(&in, nil, nil, 0, nil, 0, &out)
	if err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))

	result := make([]byte, out.Size)
	copy(result, unsafe.Slice(out.Data, out.Size))
	return result, nil
}

func decrypt(data []byte) ([]byte, error) {
	var in, out windows.DataBlob
	in.Data = &data[0]
	in.Size = uint32(len(data))

	err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, 0, &out)
	if err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))

	result := make([]byte, out.Size)
	copy(result, unsafe.Slice(out.Data, out.Size))
	return result, nil
}
