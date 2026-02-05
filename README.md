# WireGuard Mesh Builder

A Go-based tool for building and managing WireGuard mesh networks with support for NAT traversal, automatic endpoint detection, and incremental configuration updates.

## Features

- **Automatic Mesh Network Creation**: Builds full mesh topology where every node can communicate with every other node
- **NAT Detection**: Automatically detects nodes behind NAT and configures persistent keepalive
- **SSH-Based Deployment**: Installs and configures WireGuard on remote Ubuntu hosts via SSH
- **Incremental Updates**: Uses `wg set` commands for online configuration changes without restarting interfaces
- **Key Management**: Generates and stores WireGuard key pairs locally for all nodes
- **Routing Table Management**: Automatically configures routes for networks behind mesh nodes
- **Diff-Based Deployment**: Only applies configuration changes, minimizing disruption
- **Persistent State**: Stores mesh configuration in JSON format

## Prerequisites

- Go 1.23 or later
- WireGuard tools (`wg` command) on the machine running wgmesh
- SSH access to all nodes (root or sudo privileges required)
- Ubuntu-based target systems (tested on Ubuntu 20.04+)

## Installation

```bash
git clone https://github.com/ituoga/wgmeshbuilder.git
cd wgmeshbuilder
go build -o wgmesh
```

## Quick Start

### 1. Initialize a new mesh

```bash
./wgmesh -init
```

This creates a `mesh-state.json` file with default settings:
- Interface name: `wg0`
- Mesh network: `10.99.0.0/16`
- Listen port: `51820`

### 2. Add nodes to the mesh

```bash
# Format: hostname:mesh_ip:ssh_host[:ssh_port]
./wgmesh -add node1:10.99.0.1:192.168.1.10
./wgmesh -add node2:10.99.0.2:203.0.113.50
./wgmesh -add node3:10.99.0.3:198.51.100.20:2222
```

- `hostname`: Node identifier (should match the actual hostname)
- `mesh_ip`: IP address within the mesh network
- `ssh_host`: SSH connection address (can be IP or hostname)
- `ssh_port`: SSH port (optional, defaults to 22)

### 3. List nodes

```bash
./wgmesh -list
```

Output:
```
Mesh Network: 10.99.0.0/16
Interface: wg0
Listen Port: 51820

Nodes:
  node1 (local):
    Mesh IP: 10.99.0.1
    SSH: 192.168.1.10:22
    Public Key: AbCd...Ef12
    Endpoint: 192.168.1.10:51820

  node2 [NAT]:
    Mesh IP: 10.99.0.2
    SSH: 203.0.113.50:22
    Public Key: GhIj...Kl34
```

### 4. Deploy configuration

```bash
./wgmesh -deploy
```

This will:
1. Connect to each node via SSH
2. Install WireGuard if not present
3. Detect public endpoints and NAT status
4. Generate or update WireGuard configuration
5. Apply changes using `wg set` commands (if possible)
6. Configure routing tables

### 5. Remove a node

```bash
./wgmesh -remove node3
./wgmesh -deploy
```

## Advanced Usage

### Custom State File

```bash
./wgmesh -state /path/to/custom-state.json -list
```

### Adding Routes for Networks Behind Nodes

Edit the `mesh-state.json` file and add routable networks to a node:

```json
{
  "nodes": {
    "node1": {
      "hostname": "node1",
      "mesh_ip": "10.99.0.1",
      "routable_networks": ["192.168.10.0/24", "192.168.20.0/24"],
      ...
    }
  }
}
```

After editing, run `./wgmesh -deploy` to apply the changes. Other nodes in the mesh will automatically route traffic for `192.168.10.0/24` and `192.168.20.0/24` through node1.

## How It Works

### Mesh Topology

Every node becomes a peer to every other node. For a 4-node mesh:

```
node1 <----> node2
  ^            ^
  |            |
  v            v
node3 <----> node4
```

### NAT Traversal

- Nodes with public IPs are configured as endpoints for other nodes
- Nodes behind NAT use persistent keepalive to maintain connections
- The tool detects NAT by comparing SSH host with detected public IP

### Online Updates

When deploying changes, the tool:
1. Reads current WireGuard configuration using `wg show dump`
2. Calculates differences between current and desired state
3. Applies changes using `wg set` commands:
   - Add new peers
   - Update existing peers
   - Remove old peers
4. Only restarts the interface if fundamental changes are required

### SSH Key Authentication

The tool attempts authentication in this order:
1. SSH agent (if `SSH_AUTH_SOCK` is set)
2. `~/.ssh/id_rsa`
3. `~/.ssh/id_ed25519`
4. `~/.ssh/id_ecdsa`

Ensure your SSH keys are added to the `authorized_keys` file on target hosts.

## Configuration File

The `mesh-state.json` file stores the complete mesh state:

```json
{
  "interface_name": "wg0",
  "network": "10.99.0.0/16",
  "listen_port": 51820,
  "local_hostname": "control-node",
  "nodes": {
    "node1": {
      "hostname": "node1",
      "mesh_ip": "10.99.0.1",
      "public_key": "base64-encoded-public-key",
      "private_key": "base64-encoded-private-key",
      "ssh_host": "192.168.1.10",
      "ssh_port": 22,
      "listen_port": 51820,
      "public_endpoint": "192.168.1.10:51820",
      "behind_nat": false,
      "routable_networks": [],
      "is_local": true
    }
  }
}
```

## Security Considerations

- Private keys are stored in `mesh-state.json` - protect this file with appropriate permissions
- The tool uses `InsecureIgnoreHostKey` for SSH - consider implementing proper host key verification for production
- WireGuard traffic is encrypted, but the control plane (SSH) should be secured
- Root SSH access is required on target hosts

## Troubleshooting

### Connection Issues

```bash
# Test SSH connectivity
ssh root@node-address

# Check WireGuard status on a node
ssh root@node-address wg show
```

### Check Interface Status

```bash
ssh root@node-address ip addr show wg0
ssh root@node-address ip route
```

### View WireGuard Logs

```bash
ssh root@node-address journalctl -u wg-quick@wg0
```

### Rebuild Configuration

If something goes wrong, you can force a fresh configuration:
```bash
# On each node, remove the interface
ssh root@node-address ip link del wg0

# Then redeploy
./wgmesh -deploy
```

## Architecture

```
wgmeshbuilder/
├── main.go                      # CLI interface
├── pkg/
│   ├── mesh/
│   │   ├── types.go            # Data structures
│   │   ├── mesh.go             # Mesh management (add/remove/list)
│   │   └── deploy.go           # Deployment logic
│   ├── wireguard/
│   │   ├── keys.go             # Key generation
│   │   ├── config.go           # Config parsing and diffing
│   │   ├── apply.go            # Configuration application
│   │   └── convert.go          # Type conversions
│   └── ssh/
│       ├── client.go           # SSH connection management
│       └── wireguard.go        # Remote WireGuard operations
└── mesh-state.json             # Mesh state (created on init)
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - see LICENSE file for details

## Roadmap

- [ ] Support for multiple mesh networks
- [ ] Web UI for mesh management
- [ ] Monitoring and health checks
- [ ] Support for more Linux distributions
- [ ] IPv6 support
- [ ] Automatic key rotation
- [ ] Integration with service discovery systems
