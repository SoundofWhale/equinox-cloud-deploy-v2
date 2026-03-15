package security

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEncryptDecryptRoundTrip verifies AES-256-GCM encrypt/decrypt produces the original data.
func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32) // zero key is fine for testing
	for i := range key {
		key[i] = byte(i)
	}

	original := []byte("Hello, EQUINOX 2.0! This is secret media content. 🌿")

	encrypted, err := EncryptFile(key, original)
	if err != nil {
		t.Fatalf("EncryptFile failed: %v", err)
	}

	// Encrypted should be longer than original (nonce + auth tag).
	if len(encrypted) <= len(original) {
		t.Fatalf("encrypted data should be larger than original")
	}

	decrypted, err := DecryptFile(key, encrypted)
	if err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}

	if !bytes.Equal(original, decrypted) {
		t.Fatalf("round-trip mismatch:\n  original:  %q\n  decrypted: %q", original, decrypted)
	}
}

// TestEncryptDecryptWrongKey verifies that decryption with a wrong key fails.
func TestEncryptDecryptWrongKey(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	wrongKey := make([]byte, 32)
	for i := range wrongKey {
		wrongKey[i] = byte(i + 1)
	}

	encrypted, err := EncryptFile(key, []byte("secret"))
	if err != nil {
		t.Fatalf("EncryptFile failed: %v", err)
	}

	_, err = DecryptFile(wrongKey, encrypted)
	if err == nil {
		t.Fatal("expected decryption to fail with wrong key, but it succeeded")
	}
}

// TestKeyManagerParanoid verifies Argon2id key derivation and zeroing.
func TestKeyManagerParanoid(t *testing.T) {
	passphrase := []byte("my-strong-passphrase")
	salt := []byte("equinox-salt-16b")

	km := NewKeyManagerParanoid(passphrase, salt)

	// Key should be 32 bytes.
	if len(km.RawKey()) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(km.RawKey()))
	}

	// HexKey should be 64 chars.
	if len(km.HexKey()) != 64 {
		t.Fatalf("expected 64-char hex key, got %d", len(km.HexKey()))
	}

	// Same passphrase + salt should give deterministic key.
	km2 := NewKeyManagerParanoid(passphrase, salt)
	if km.HexKey() != km2.HexKey() {
		t.Fatal("same passphrase+salt should produce same key")
	}

	// After ZeroKey, all bytes should be 0.
	km.ZeroKey()
	for i, b := range km.RawKey() {
		if b != 0 {
			t.Fatalf("key byte %d not zeroed: %d", i, b)
		}
	}
}

// TestGenerateMnemonic verifies BIP-39 mnemonic generation produces 12 words.
func TestGenerateMnemonic(t *testing.T) {
	mnemonic, err := GenerateMnemonic()
	if err != nil {
		t.Fatalf("GenerateMnemonic failed: %v", err)
	}

	words := strings.Fields(mnemonic)
	if len(words) != 12 {
		t.Fatalf("expected 12 words, got %d: %q", len(words), mnemonic)
	}

	// Each mnemonic should be different (probabilistic — generate two).
	mnemonic2, _ := GenerateMnemonic()
	if mnemonic == mnemonic2 {
		t.Fatal("two consecutive mnemonics should differ")
	}
}

// TestDeriveKeyFromMnemonic checks deterministic key derivation from mnemonic.
func TestDeriveKeyFromMnemonic(t *testing.T) {
	mnemonic, _ := GenerateMnemonic()

	key1, err := DeriveKeyFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("DeriveKeyFromMnemonic failed: %v", err)
	}
	if len(key1) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key1))
	}

	// Same mnemonic should produce same key.
	key2, _ := DeriveKeyFromMnemonic(mnemonic)
	if !bytes.Equal(key1, key2) {
		t.Fatal("same mnemonic should produce same key")
	}
}

// TestDeriveKeyFromMnemonicInvalid checks that invalid mnemonic returns error.
func TestDeriveKeyFromMnemonicInvalid(t *testing.T) {
	_, err := DeriveKeyFromMnemonic("this is not a valid bip39 mnemonic at all nope")
	if err == nil {
		t.Fatal("expected error for invalid mnemonic")
	}
}

// TestOpenDB verifies that SQLCipher DB opens, migrates, and reports cipher info.
// This test creates a temp DB file — no Docker required.
func TestOpenDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_vault.db")

	// Use a test key (32 bytes = 64 hex chars).
	km := NewKeyManagerParanoid([]byte("test"), []byte("salt-16-bytes!!!"))

	db, err := OpenDB(dbPath, km.HexKey())
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer db.Close()

	// Verify cipher version is reported.
	ver, err := db.CipherVersion()
	if err != nil {
		t.Fatalf("CipherVersion failed: %v", err)
	}
	if ver == "" {
		t.Fatal("CipherVersion returned empty string")
	}
	t.Logf("SQLCipher version: %s", ver)

	// Verify cipher page size.
	pageSize, err := db.CipherPageSize()
	if err != nil {
		t.Fatalf("CipherPageSize failed: %v", err)
	}
	t.Logf("cipher_page_size: %d", pageSize)

	// Verify tables exist by inserting a slot.
	_, err = db.Conn.Exec(`INSERT INTO slots (id, name, start_time, end_time, recur_rule, slot_type)
		VALUES ('test-1', 'Test Slot', '09:00', '10:00', 'daily', 'hard')`)
	if err != nil {
		t.Fatalf("insert into slots failed (table missing?): %v", err)
	}

	// Verify DB file was created on disk.
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("DB file was not created on disk")
	}
}
