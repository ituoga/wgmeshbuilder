#!/bin/bash
# Test script to demonstrate encryption functionality

echo "=== WireGuard Mesh Builder - Encryption Test ==="
echo ""
echo "This script demonstrates the encryption feature."
echo ""

cat << 'EOF'
## Usage Examples

### 1. Initialize with encryption (creates encrypted state file)
$ ./wgmesh --encrypt -init
Enter encryption password: ********
Confirm password: ********
Mesh initialized successfully

The mesh-state.json file will contain base64-encoded encrypted data like:
Q1RZNlBXMTJHVzR4TVRrMllXNWpaVzkxZEdWd1pYSmhkR2x2Ym...

### 2. Add node with encryption
$ ./wgmesh --encrypt --add node1:10.99.0.1:192.168.1.10
Enter encryption password: ********
Added node: node1 (10.99.0.1)
Node added successfully

### 3. List nodes with encryption
$ ./wgmesh --encrypt --list
Enter encryption password: ********
Mesh Network: 10.99.0.0/16
Interface: wg0
Listen Port: 51820

Nodes:
  node1:
    Mesh IP: 10.99.0.1
    ...

### 4. Deploy with encryption
$ ./wgmesh --encrypt --deploy
Enter encryption password: ********
Detecting endpoints...
Deploying to node1...
  ...

### 5. Without --encrypt flag (on encrypted file)
$ ./wgmesh --list
Failed to load mesh state: failed to parse state file: invalid character 'Q' looking for beginning of value

This fails because the file is encrypted and needs the password!

## How Encryption Works

1. **Password-based encryption**: Uses PBKDF2 to derive a 256-bit key from your password
2. **AES-256-GCM**: Industry-standard authenticated encryption
3. **Random salt**: Each encryption uses a unique 32-byte salt
4. **Base64 encoding**: Encrypted data is base64-encoded for easy storage in vaults

## Security Features

- Salt: 32 bytes random
- Key derivation: PBKDF2 with 100,000 iterations and SHA-256
- Encryption: AES-256-GCM (authenticated encryption)
- Nonce: Random per encryption
- Output: Base64-encoded (vault-friendly)

## File Format Comparison

### Without encryption (mesh-state.json):
{
  "interface_name": "wg0",
  "network": "10.99.0.0/16",
  "nodes": {
    "node1": {
      "private_key": "yLmO9xPq..."
    }
  }
}

### With encryption (mesh-state.json):
U2FsdGVkX1+Qq1RZNlBXMTJHVzR4TVRrMllXNWpaVzkxZEdWd0FsSnZibk5oY0dWaGRHbHZi
bm1KekxYQkhjM04zYjNKa0lqb2dJbTFsYzJndGMzUmhkR1V1YW5OdmJpSXNJQ0p1WlhSM2Iz
...
(long base64 string)

## Storing in Vault

Since the encrypted file is base64-encoded, you can easily store it in:
- HashiCorp Vault
- AWS Secrets Manager
- Azure Key Vault
- Environment variables
- Git (with caution - encrypted but still visible)

Example for HashiCorp Vault:
$ vault kv put secret/wgmesh state=@mesh-state.json
$ vault kv get -field=state secret/wgmesh > mesh-state.json
$ ./wgmesh --encrypt --list

EOF

echo ""
echo "=== To actually test encryption, run these commands interactively: ==="
echo ""
echo "# Clean up any existing state"
echo "rm -f mesh-state.json"
echo ""
echo "# Initialize with encryption (will prompt for password)"
echo "./wgmesh --encrypt -init"
echo ""
echo "# View the encrypted file"
echo "cat mesh-state.json"
echo ""
echo "# List nodes (will prompt for password)"
echo "./wgmesh --encrypt --list"
