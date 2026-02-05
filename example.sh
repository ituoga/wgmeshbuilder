#!/bin/bash
# Example usage of wgmesh

set -e

echo "=== WireGuard Mesh Builder Example ==="
echo ""

echo "1. Initialize mesh"
./wgmesh -init
echo ""

echo "2. Add nodes"
./wgmesh -add server1:10.99.0.1:192.168.1.10
./wgmesh -add server2:10.99.0.2:192.168.1.20
./wgmesh -add server3:10.99.0.3:192.168.1.30
echo ""

echo "3. List nodes"
./wgmesh -list
echo ""

echo "4. (Optional) Deploy to nodes"
echo "   Run: ./wgmesh -deploy"
echo "   This will connect to each node via SSH and configure WireGuard"
echo ""

echo "5. (Optional) Remove a node"
echo "   Run: ./wgmesh -remove server3"
echo "   Then: ./wgmesh -deploy"
echo ""

echo "=== Example complete ==="
echo "State saved in mesh-state.json"
