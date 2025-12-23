//go:build !windows

package auth

import (
	"fmt"
	"os"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func persistKey(path string, privKey wgtypes.Key) error {
	// 0600: read/write by owner only
	err := os.WriteFile(path, []byte(privKey.String()), 0600)
	if err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}
	return nil
}

func loadKey(path string) (wgtypes.Key, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return wgtypes.Key{}, err
	}

	privKey, err := wgtypes.ParseKey(string(data))
	if err != nil {
		return wgtypes.Key{}, fmt.Errorf("failed to parse key: %w", err)
	}

	return privKey, nil
}
