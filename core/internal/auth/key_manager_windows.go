//go:build windows

package auth

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func persistKey(path string, privKey wgtypes.PrivateKey) error {
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

func loadKey(path string) (wgtypes.PrivateKey, error) {
	encrypted, err := os.ReadFile(path)
	if err != nil {
		return wgtypes.PrivateKey{}, err
	}

	// Decrypt using DPAPI
	decrypted, err := decrypt(encrypted)
	if err != nil {
		return wgtypes.PrivateKey{}, fmt.Errorf("failed to decrypt key with DPAPI: %w", err)
	}

	privKey, err := wgtypes.ParseKey(string(decrypted))
	if err != nil {
		return wgtypes.PrivateKey{}, fmt.Errorf("failed to parse key: %w", err)
	}

	return wgtypes.PrivateKey(privKey), nil
}

func encrypt(data []byte) ([]byte, error) {
	var in, out windows.DataBlob
	in.Data = &data[0]
	in.Size = uint32(len(data))

	err := windows.CryptProtectData(&in, nil, nil, 0, nil, 0, &out)
	if err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(out.Data))

	result := make([]byte, out.Size)
	copy(result, out.DataToSlice())
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
	defer windows.LocalFree(windows.Handle(out.Data))

	result := make([]byte, out.Size)
	copy(result, out.DataToSlice())
	return result, nil
}
