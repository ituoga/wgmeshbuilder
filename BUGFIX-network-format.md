# Bug Fix: Network Format Normalization

## Problem

When reading current routes from `ip route show`, Linux displays host routes (single IPs) without the `/32` CIDR notation:

```bash
$ ip route show dev wg0
192.168.5.5 via 10.199.0.1       # Note: no /32
192.168.10.0/24 via 10.199.0.1
```

But when we add routes or compare with desired state, we use full CIDR notation:
```
192.168.5.5/32 via 10.199.0.1
```

This caused the diff algorithm to think these were different routes:
- Current: `192.168.5.5` (no /32)
- Desired: `192.168.5.5/32` (with /32)
- Result: **Incorrectly removed and re-added the same route!**

## Solution

Added `normalizeNetwork()` function that:
1. Detects if a network has no CIDR notation (`/`)
2. For IPv4 addresses (3 dots), adds `/32`
3. For IPv6 addresses (contains `:`), adds `/128`
4. Leaves networks with existing CIDR notation unchanged

## Before Fix

```
  Removing stale routes...
    Removed route: 192.168.5.5 via 10.199.0.1
  Adding new routes...
    Added route: 192.168.5.5/32 via 10.199.0.1
```
*Unnecessary removal and re-add of the same route*

## After Fix

```
  No route changes needed (all routes already correct)
```
*Correctly detects that 192.168.5.5 and 192.168.5.5/32 are the same route*

## Test Cases

### Test 1: Single IP (/32 route)
```
Current from ip route: "192.168.5.5 via 10.199.0.1"
After normalization:   "192.168.5.5/32 via 10.199.0.1"
Desired:              "192.168.5.5/32 via 10.199.0.1"
Match: ✓ (no changes)
```

### Test 2: Network with CIDR
```
Current from ip route: "192.168.10.0/24 via 10.199.0.1"
After normalization:   "192.168.10.0/24 via 10.199.0.1"
Desired:              "192.168.10.0/24 via 10.199.0.1"
Match: ✓ (no changes)
```

### Test 3: IPv6 single address
```
Current from ip route: "2001:db8::1 via fe80::1"
After normalization:   "2001:db8::1/128 via fe80::1"
Desired:              "2001:db8::1/128 via fe80::1"
Match: ✓ (no changes)
```

## Implementation

```go
func normalizeNetwork(network string) string {
    // If network doesn't contain '/', it's a host route
    if !strings.Contains(network, "/") {
        // IPv4: add /32
        if strings.Count(network, ".") == 3 {
            return network + "/32"
        }
        // IPv6: add /128
        if strings.Contains(network, ":") {
            return network + "/128"
        }
    }
    return network
}
```

Called when parsing routes from `ip route show dev wg0`.
