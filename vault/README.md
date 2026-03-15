# Vault — Encrypted Media Storage

This directory contains **encrypted media files** managed by The Warden agent.

## Structure

```
vault/
└── media/     ← All user media files — each encrypted with AES-256-GCM
```

## Important Rules (Enforced by The Warden)

1. **Never** write plaintext files here directly.
2. Every file is encrypted before write using a unique AES-256-GCM nonce.
3. The nonce is prepended to the ciphertext blob.
4. The filename stored here is a UUID — metadata (original name, MIME type) lives in the SQLCipher DB.
5. This directory is mounted as a Docker volume (`vault_data`) — it persists across container restarts.

## Reference

See [equinox_warden_skills.md](file:///c:/AI/skills/equinox_warden_skills.md) → Section 1: Media File Encryption
