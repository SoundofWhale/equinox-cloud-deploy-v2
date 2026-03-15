package security

import (
	"encoding/hex"
	"fmt"

	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/argon2"
)

const (
	keychainService = "equinox"
	keychainUser    = "master_key"

	// Argon2id parameters (from spec)
	argonTime    = 3
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
	argonKeyLen  = 32 // 256 bits
)

// KeyMode determines how the master key is stored.
type KeyMode int

const (
	ModeParanoid    KeyMode = iota // RAM only — zeroed on exit
	ModeConvenience                // OS Keychain via go-keyring
)

// KeyManager handles master key derivation, storage, and cleanup.
type KeyManager struct {
	Mode KeyMode
	key  []byte // 32-byte raw key — always in memory while active
}

// NewKeyManagerParanoid creates a KeyManager in Paranoid mode.
// The key is derived from passphrase+salt using Argon2id and held only in RAM.
func NewKeyManagerParanoid(passphrase, salt []byte) *KeyManager {
	key := argon2.IDKey(passphrase, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return &KeyManager{Mode: ModeParanoid, key: key}
}

// NewKeyManagerConvenience creates a KeyManager that loads the key from the OS Keychain.
// If no key exists in the keychain, it returns an error.
func NewKeyManagerConvenience() (*KeyManager, error) {
	hexKey, err := keyring.Get(keychainService, keychainUser)
	if err != nil {
		return nil, fmt.Errorf("keychain: %w", err)
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("keychain hex decode: %w", err)
	}
	return &KeyManager{Mode: ModeConvenience, key: key}, nil
}

// SaveToKeychain stores the current key in the OS Keychain (Convenience mode only).
func (km *KeyManager) SaveToKeychain() error {
	return keyring.Set(keychainService, keychainUser, hex.EncodeToString(km.key))
}

// HexKey returns the 32-byte key encoded as a 64-char hex string
// (used as PRAGMA key for SQLCipher).
func (km *KeyManager) HexKey() string {
	return hex.EncodeToString(km.key)
}

// RawKey returns the raw 32-byte key slice (for AES-256 encryption).
func (km *KeyManager) RawKey() []byte {
	return km.key
}

// ZeroKey securely wipes the in-memory key by overwriting every byte with 0.
func (km *KeyManager) ZeroKey() {
	for i := range km.key {
		km.key[i] = 0
	}
}

// Close zeroes the key. In Paranoid mode, the key is gone forever after this call.
func (km *KeyManager) Close() {
	km.ZeroKey()
}
