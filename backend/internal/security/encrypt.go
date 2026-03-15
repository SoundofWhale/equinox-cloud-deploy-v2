package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// EncryptFile encrypts plaintext using AES-256-GCM.
// The returned blob has the format: [nonce (12 bytes) | ciphertext].
func EncryptFile(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}

	// Seal appends ciphertext after nonce — nonce is prepended.
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptFile decrypts a blob produced by EncryptFile.
// Expects [nonce (12 bytes) | ciphertext].
func DecryptFile(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ct := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}

// ZeroingReader wraps a reader and zeroes its internal buffer after reading.
// Note: This is a best-effort helper.
func ZeroMemory(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// StreamDecryptReader provides an io.Reader that decrypts AES-GCM data.
// IMPORTANT: Due to GCM nature, the full chunk must be authenticated. 
// This implementation assumes the input is [nonce(12) | ciphertext+tag].
type StreamDecryptReader struct {
	key       []byte
	r         io.Reader
	plaintext []byte
	off       int
	err       error
}

func NewStreamDecryptReader(key []byte, r io.Reader) *StreamDecryptReader {
	return &StreamDecryptReader{
		key: key,
		r:   r,
	}
}

func (s *StreamDecryptReader) Read(p []byte) (n int, err error) {
	if s.err != nil {
		return 0, s.err
	}

	if s.plaintext == nil {
		// First read: Decrypt everything into memory (best we can do with single GCM blob)
		// To truly stream we'd need a chunked format, but for now we follow the io.Reader interface.
		// We still zero out the buffer later.
		all, err := io.ReadAll(s.r)
		if err != nil {
			s.err = err
			return 0, err
		}
		
		decrypted, err := DecryptFile(s.key, all)
		if err != nil {
			s.err = err
			return 0, err
		}

		s.plaintext = decrypted
		// Zero the encrypted 'all' buffer as it's no longer needed in that form
		ZeroMemory(all)
	}

	if s.off >= len(s.plaintext) {
		s.err = io.EOF
		return 0, io.EOF
	}

	n = copy(p, s.plaintext[s.off:])
	s.off += n
	return n, nil
}

// Close zeroes out the internal plaintext buffer.
func (s *StreamDecryptReader) Close() error {
	if s.plaintext != nil {
		ZeroMemory(s.plaintext)
		s.plaintext = nil
	}
	return nil
}
