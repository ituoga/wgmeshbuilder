package ssh

import (
	"fmt"
	"strings"
)

func EnsureWireGuardInstalled(client *Client) error {
	output, err := client.Run("which wg")
	if err == nil && strings.Contains(output, "/wg") {
		return nil
	}

	fmt.Println("  Installing WireGuard...")

	commands := []string{
		"apt-get update -qq",
		"DEBIAN_FRONTEND=noninteractive apt-get install -y -qq wireguard wireguard-tools",
		"modprobe wireguard || true",
	}

	for _, cmd := range commands {
		if _, err := client.Run(cmd); err != nil {
			return fmt.Errorf("failed to run %q: %w", cmd, err)
		}
	}

	return nil
}

func DetectPublicIP(client *Client) (string, error) {
	output, err := client.Run("curl -s -4 ifconfig.me || curl -s -4 icanhazip.com || true")
	if err != nil {
		return "", err
	}

	ip := strings.TrimSpace(output)
	if ip == "" {
		return "", fmt.Errorf("could not detect public IP")
	}

	return ip, nil
}

func UpdateRoutingTable(client *Client, iface string, networks []string) error {
	fmt.Printf("  Updating routing table...\n")

	for _, network := range networks {
		cmd := fmt.Sprintf("ip route add %s dev %s || ip route replace %s dev %s",
			network, iface, network, iface)
		if err := client.RunQuiet(cmd); err != nil {
			return fmt.Errorf("failed to add route for %s: %w", network, err)
		}
		fmt.Printf("    Added route: %s\n", network)
	}

	cmd := "sysctl -w net.ipv4.ip_forward=1 > /dev/null"
	if err := client.RunQuiet(cmd); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %w", err)
	}

	return nil
}
