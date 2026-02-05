# State File Encryption

WireGuard Mesh Builder supports encrypting the mesh state file to protect sensitive data (WireGuard private keys, network topology, etc.).

## Quick Start

```bash
# Initialize with encryption
./wgmesh --encrypt -init

# All operations require --encrypt flag and password
./wgmesh --encrypt --add node1:10.99.0.1:192.168.1.10
./wgmesh --encrypt --list
./wgmesh --encrypt --deploy
```

## Why Encrypt?

The `mesh-state.json` file contains:
- WireGuard private keys for all nodes
- Network topology and IP addresses
- SSH connection details
- Routing configuration

**Without encryption**, this file must be protected with strict file permissions and cannot be safely stored in version control or shared vaults.

**With encryption**, the file is AES-256-GCM encrypted and base64-encoded, making it safe to:
- Store in HashiCorp Vault, AWS Secrets Manager, etc.
- Back up to cloud storage
- Share between team members (with password via separate channel)

## How It Works

### Encryption Process

1. **Password Input**: User enters password (confirmed twice on init)
2. **Key Derivation**: PBKDF2 with 100,000 iterations and SHA-256 derives 256-bit key
3. **Random Salt**: 32-byte random salt generated per encryption
4. **Encryption**: AES-256-GCM encrypts the JSON data
5. **Encoding**: Result is base64-encoded for text-safe storage

### Decryption Process

1. **Password Input**: User enters password
2. **Decoding**: Base64-decode the file content
3. **Salt Extraction**: Extract 32-byte salt from beginning
4. **Key Derivation**: PBKDF2 derives key from password + salt
5. **Decryption**: AES-256-GCM decrypts and verifies authenticity
6. **Parsing**: JSON is parsed into mesh state

## Security Properties

### Encryption Algorithm
- **Cipher**: AES-256 (Advanced Encryption Standard with 256-bit key)
- **Mode**: GCM (Galois/Counter Mode) - provides both confidentiality and authenticity
- **Authentication**: Built-in authentication tag prevents tampering

### Key Derivation
- **Function**: PBKDF2 (Password-Based Key Derivation Function 2)
- **Hash**: SHA-256
- **Iterations**: 100,000 (protects against brute-force attacks)
- **Salt**: 32 bytes random (prevents rainbow table attacks)
- **Key Size**: 256 bits

### Random Values
- **Salt**: 32 bytes per encryption (unique for each save)
- **Nonce**: 12 bytes per encryption (GCM requirement)
- **Source**: `crypto/rand` (cryptographically secure)

## File Format

### Unencrypted (plain JSON)
```json
{
  "interface_name": "wg0",
  "network": "10.99.0.0/16",
  "listen_port": 51820,
  "nodes": {
    "node1": {
      "hostname": "node1",
      "private_key": "yLmO9xPq4r5s6t7u8v9w0x1y2z3a4b5c6d7e8f9g0h1i2j3k=",
      ...
    }
  }
}
```

### Encrypted (base64-encoded ciphertext)
```
U2FsdGVkX1+Qq1RZNlBXMTJHVzR4TVRrMllXNWpaVzkxZEdWd0FsSnZibk5hY0dWaGRHbHZi
bm1KekxYQkhjM04zYjNKa0lqb2dJbTFsYzJndGMzUmhkR1V1YW5OdmJpSXNJQ0p1WlhSM2Iz
SjJJam9nSWpFd0xqazVMakF1TUM4eE5pSXNJQ0pzYVhOMFpXNWZjRzl5ZENJNklEVXhPREl3
...
(long base64 string, ~2-3x size of original JSON)
```

### Binary Structure
```
[Salt: 32 bytes][Nonce: 12 bytes][Ciphertext: variable][Auth Tag: 16 bytes]
```

## Usage Examples

### Initialize Encrypted Mesh

```bash
$ ./wgmesh --encrypt -init
Enter encryption password: ********
Confirm password: ********
Mesh initialized successfully

$ cat mesh-state.json
U2FsdGVkX1+Qq1RZNlBXMTJH...
```

### Add Node

```bash
$ ./wgmesh --encrypt --add node1:10.99.0.1:192.168.1.10
Enter encryption password: ********
Added node: node1 (10.99.0.1)
Node added successfully
```

### List Nodes

```bash
$ ./wgmesh --encrypt --list
Enter encryption password: ********
Mesh Network: 10.99.0.0/16
Interface: wg0
Listen Port: 51820

Nodes:
  node1:
    Mesh IP: 10.99.0.1
    SSH: 192.168.1.10:22
    ...
```

### Deploy

```bash
$ ./wgmesh --encrypt --deploy
Enter encryption password: ********
Detecting endpoints...
Deploying to node1...
  ✓ Deployed successfully
```

### Wrong Password

```bash
$ ./wgmesh --encrypt --list
Enter encryption password: ********
Failed to load mesh state: failed to decrypt state file: decryption failed (wrong password?): cipher: message authentication failed
```

## Vault Integration

### HashiCorp Vault

```bash
# Store encrypted state in Vault
vault kv put secret/wgmesh state=@mesh-state.json

# Retrieve from Vault
vault kv get -field=state secret/wgmesh > mesh-state.json

# Use with wgmesh
./wgmesh --encrypt --list
```

### AWS Secrets Manager

```bash
# Store
aws secretsmanager create-secret \
  --name wgmesh-state \
  --secret-string file://mesh-state.json

# Retrieve
aws secretsmanager get-secret-value \
  --secret-id wgmesh-state \
  --query SecretString \
  --output text > mesh-state.json

# Use
./wgmesh --encrypt --list
```

### Azure Key Vault

```bash
# Store
az keyvault secret set \
  --vault-name myvault \
  --name wgmesh-state \
  --file mesh-state.json

# Retrieve
az keyvault secret show \
  --vault-name myvault \
  --name wgmesh-state \
  --query value -o tsv > mesh-state.json

# Use
./wgmesh --encrypt --list
```

## Best Practices

### Password Management

1. **Strong Passwords**: Use at least 20 characters with mixed case, numbers, and symbols
2. **Password Managers**: Generate and store password in a password manager
3. **Separate Channel**: Share password via different channel than encrypted file
4. **Rotation**: Consider rotating password periodically

### File Storage

1. **Backup**: Keep encrypted backups in multiple locations
2. **Version Control**: Safe to commit encrypted file to git (password separate)
3. **Access Control**: Limit who has the encryption password
4. **Audit**: Log access to encrypted state files

### Operational Security

1. **No Plaintext Copies**: Delete unencrypted state files after encrypting
2. **Memory**: Password is only stored in memory during operation
3. **Logs**: Passwords never logged or written to disk
4. **Terminal History**: Use space before command to prevent history: ` ./wgmesh --encrypt --list`

## Performance

Encryption operations are fast:
- **Encryption**: ~1-2ms for typical mesh state (<100KB)
- **Decryption**: ~1-2ms
- **Key Derivation**: ~100ms (intentionally slow to resist brute-force)

## Limitations

1. **Password Required**: Every operation needs password (no cached sessions)
2. **No Key Rotation**: Changing password requires decrypting and re-encrypting
3. **All-or-Nothing**: Cannot encrypt only parts of the state file
4. **Single Password**: All operations use same password (no per-user passwords)

## Migration

### From Unencrypted to Encrypted

```bash
# Backup original
cp mesh-state.json mesh-state.json.backup

# Load unencrypted, save encrypted
./wgmesh --list  # Works with unencrypted
./wgmesh --encrypt --list  # Will prompt for password and re-save encrypted

# Verify
cat mesh-state.json  # Should now be base64-encoded
./wgmesh --encrypt --list  # Should work with password
./wgmesh --list  # Should fail (file is now encrypted)

# Delete backup once confirmed
rm mesh-state.json.backup
```

### From Encrypted to Unencrypted

```bash
# Backup encrypted version
cp mesh-state.json mesh-state.json.encrypted

# Load encrypted, save unencrypted
./wgmesh --encrypt --list  # Loads and decrypts
./wgmesh --list  # Re-saves as unencrypted

# Verify
cat mesh-state.json  # Should now be readable JSON

# Keep encrypted backup
mv mesh-state.json.encrypted ~/secure-backup/
```

## Troubleshooting

### "Failed to decrypt: wrong password"
- Verify password is correct
- Check if file is corrupted (compare with backup)

### "Failed to decode base64"
- File may not be encrypted
- Try without `--encrypt` flag

### "Invalid character looking for beginning of value"
- File is encrypted but you forgot `--encrypt` flag
- Add `--encrypt` and provide password

## Implementation Details

See source files:
- `pkg/crypto/encrypt.go` - Encryption/decryption logic
- `pkg/crypto/password.go` - Password input handling
- `pkg/mesh/mesh.go` - Integration with mesh state
