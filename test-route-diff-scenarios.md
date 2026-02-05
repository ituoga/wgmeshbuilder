# Route Diff Optimization Test Scenarios

## How the optimized diff works

The route diff algorithm now uses a smart key-based matching system:
- **Key format**: `"network|gateway"` (e.g., `"192.168.10.0/24|10.99.0.1"`)
- **Exact matching**: Routes are only touched if they actually changed
- **Gateway tracking**: Detects when the same network changes to a different mesh node

## Test Scenarios

### Scenario 1: No Changes (Optimization in Action)
**Current state:**
```
192.168.10.0/24 via 10.99.0.1 dev wg0
192.168.20.0/24 via 10.99.0.1 dev wg0
```

**Desired state:**
```
192.168.10.0/24 via 10.99.0.1
192.168.20.0/24 via 10.99.0.1
```

**Result:**
- toAdd: []
- toRemove: []
- **No operations performed** ✓ (This is the optimization!)

---

### Scenario 2: Add New Network
**Current state:**
```
192.168.10.0/24 via 10.99.0.1 dev wg0
```

**Desired state:**
```
192.168.10.0/24 via 10.99.0.1
192.168.20.0/24 via 10.99.0.1
```

**Result:**
- toAdd: [192.168.20.0/24 via 10.99.0.1]
- toRemove: []
- **Only adds the new route** ✓

---

### Scenario 3: Remove Network
**Current state:**
```
192.168.10.0/24 via 10.99.0.1 dev wg0
192.168.20.0/24 via 10.99.0.1 dev wg0
```

**Desired state:**
```
192.168.10.0/24 via 10.99.0.1
```

**Result:**
- toAdd: []
- toRemove: [192.168.20.0/24 via 10.99.0.1]
- **Only removes the deleted route** ✓

---

### Scenario 4: Change Gateway (VPN IP change)
**Current state:**
```
192.168.10.0/24 via 10.99.0.1 dev wg0
```

**Desired state:**
```
192.168.10.0/24 via 10.99.0.2  (moved to different node)
```

**Result:**
- toRemove: [192.168.10.0/24 via 10.99.0.1]
- toAdd: [192.168.10.0/24 via 10.99.0.2]
- **Removes old gateway, adds new gateway** ✓

---

### Scenario 5: Mixed Changes
**Current state:**
```
192.168.10.0/24 via 10.99.0.1 dev wg0
192.168.20.0/24 via 10.99.0.1 dev wg0
192.168.30.0/24 via 10.99.0.2 dev wg0
```

**Desired state:**
```
192.168.10.0/24 via 10.99.0.1  (unchanged)
192.168.20.0/24 via 10.99.0.3  (gateway changed)
192.168.40.0/24 via 10.99.0.2  (new network)
```

**Result:**
- toRemove: [192.168.20.0/24 via 10.99.0.1, 192.168.30.0/24 via 10.99.0.2]
- toAdd: [192.168.20.0/24 via 10.99.0.3, 192.168.40.0/24 via 10.99.0.2]
- 192.168.10.0/24 via 10.99.0.1 is **NOT touched** ✓

---

## Real-World Example

**Initial deployment:**
```bash
# node1 has: routable_networks: ["192.168.10.0/24"]
# node2 has: routable_networks: ["192.168.20.0/24"]
./wgmesh -deploy
```

**On node3, routes added:**
```
ip route add 192.168.10.0/24 via 10.99.0.1 dev wg0
ip route add 192.168.20.0/24 via 10.99.0.2 dev wg0
```

**Later, no changes made:**
```bash
./wgmesh -deploy  # Re-deploy without changes
```

**On node3:**
- Reads current routes: finds 192.168.10.0/24 via 10.99.0.1, 192.168.20.0/24 via 10.99.0.2
- Calculates diff: No differences found
- Output: "No route changes needed"
- **No ip route commands executed** ✓

**Then, move network from node1 to node2:**
```bash
# Edit mesh-state.json:
# node1: routable_networks: []
# node2: routable_networks: ["192.168.10.0/24", "192.168.20.0/24"]
./wgmesh -deploy
```

**On node3:**
- Removes: ip route del 192.168.10.0/24 via 10.99.0.1 dev wg0
- Adds: ip route add 192.168.10.0/24 via 10.99.0.2 dev wg0
- **Does NOT touch** 192.168.20.0/24 via 10.99.0.2 (already correct) ✓

---

## Key Optimizations

1. **Exact matching by network+gateway**: Prevents unnecessary removals and re-adds
2. **Gateway change detection**: Explicitly handles when a network moves between mesh nodes
3. **No-op on identical state**: If routes are already correct, no commands are executed
4. **Selective removal**: Only removes routes that are truly obsolete or need gateway change
