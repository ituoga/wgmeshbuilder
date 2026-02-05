#!/bin/bash
# Test script to demonstrate route cleanup functionality

set -e

echo "=== Testing Route Cleanup ==="
echo ""

# Initialize mesh
echo "1. Initialize mesh"
./wgmesh -init
echo ""

# Add nodes
echo "2. Add test nodes"
./wgmesh -add node1:10.99.0.1:192.168.1.10
./wgmesh -add node2:10.99.0.2:192.168.1.20
./wgmesh -add node3:10.99.0.3:192.168.1.30
echo ""

# Manually edit the state file to add routable networks
echo "3. Adding routable networks to node1"
cat > mesh-state.json << 'EOF'
{
  "interface_name": "wg0",
  "network": "10.99.0.0/16",
  "listen_port": 51820,
  "local_hostname": "$(hostname)",
  "nodes": {
    "node1": {
      "hostname": "node1",
      "mesh_ip": "10.99.0.1",
      "public_key": "...",
      "private_key": "...",
      "ssh_host": "192.168.1.10",
      "ssh_port": 22,
      "listen_port": 51820,
      "behind_nat": false,
      "routable_networks": ["192.168.10.0/24", "192.168.20.0/24"],
      "is_local": false
    },
    "node2": {
      "hostname": "node2",
      "mesh_ip": "10.99.0.2",
      "public_key": "...",
      "private_key": "...",
      "ssh_host": "192.168.1.20",
      "ssh_port": 22,
      "listen_port": 51820,
      "behind_nat": false,
      "routable_networks": [],
      "is_local": false
    },
    "node3": {
      "hostname": "node3",
      "mesh_ip": "10.99.0.3",
      "public_key": "...",
      "private_key": "...",
      "ssh_host": "192.168.1.30",
      "ssh_port": 22,
      "listen_port": 51820,
      "behind_nat": false,
      "routable_networks": [],
      "is_local": false
    }
  }
}
EOF

echo "State file now has node1 with routable_networks: [192.168.10.0/24, 192.168.20.0/24]"
echo ""

echo "4. When you deploy (./wgmesh -deploy):"
echo "   - node2 will get: ip route add 192.168.10.0/24 via 10.99.0.1 dev wg0"
echo "   - node2 will get: ip route add 192.168.20.0/24 via 10.99.0.1 dev wg0"
echo "   - node3 will get the same routes"
echo ""

echo "5. If you remove 192.168.20.0/24 from node1's routable_networks and redeploy:"
echo "   - node2 will remove: ip route del 192.168.20.0/24 via 10.99.0.1 dev wg0"
echo "   - node3 will remove the same route"
echo "   - But 192.168.10.0/24 will remain"
echo ""

echo "6. The persistent config (/etc/wireguard/wg0.conf) will also be updated automatically"
echo ""

echo "=== Test Setup Complete ==="
echo "Note: This is a demonstration script. Actual deployment requires SSH access to nodes."
echo "Clean up with: rm mesh-state.json"
