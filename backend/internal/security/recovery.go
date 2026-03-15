package security

import (
	"crypto/sha256"
	"fmt"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/argon2"
)

// GenerateMnemonic creates a new 12-word BIP-39 mnemonic (128-bit entropy).
// This should be shown to the user ONCE at registration and never stored in plaintext.
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", fmt.Errorf("entropy: %w", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("mnemonic: %w", err)
	}
	return mnemonic, nil
}

// DeriveKeyFromMnemonic derives a 32-byte AES-256 key from a BIP-39 mnemonic.
// Uses the BIP-39 seed as input to Argon2id for extra hardening.
func DeriveKeyFromMnemonic(mnemonic string) ([]byte, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	// BIP-39 seed (64 bytes) from mnemonic with empty passphrase.
	seed := bip39.NewSeed(mnemonic, "")

	// Use SHA-256 of the seed as Argon2id salt for deterministic derivation.
	saltHash := sha256.Sum256(seed[:32])
	salt := saltHash[:16]

	key := argon2.IDKey(seed, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return key, nil
}
