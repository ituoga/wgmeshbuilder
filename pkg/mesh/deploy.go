package mesh

import (
	"fmt"

	"github.com/ituoga/wgmeshbuilder/pkg/ssh"
	"github.com/ituoga/wgmeshbuilder/pkg/wireguard"
)

type WireGuardConfig = wireguard.FullConfig
type WGInterface = wireguard.WGInterface
type WGPeer = wireguard.WGPeer

func (m *Mesh) Deploy() error {
	if err := m.detectEndpoints(); err != nil {
		return fmt.Errorf("failed to detect endpoints: %w", err)
	}

	for hostname, node := range m.Nodes {
		fmt.Printf("Deploying to %s...\n", hostname)

		client, err := ssh.NewClient(node.SSHHost, node.SSHPort)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", hostname, err)
		}
		defer client.Close()

		if err := ssh.EnsureWireGuardInstalled(client); err != nil {
			return fmt.Errorf("failed to ensure WireGuard on %s: %w", hostname, err)
		}

		config := m.generateConfigForNode(node)
		routes := m.collectRoutesForNode(node)

		currentConfig, err := wireguard.GetCurrentConfig(client, m.InterfaceName)
		if err != nil {
			fmt.Printf("  No existing config, applying fresh persistent configuration\n")
			if err := wireguard.ApplyPersistentConfig(client, m.InterfaceName, config, routes); err != nil {
				return fmt.Errorf("failed to apply config to %s: %w", hostname, err)
			}
		} else {
			diff := wireguard.CalculateDiff(currentConfig, wireguard.FullConfigToConfig(config))
			if diff.HasChanges() {
				fmt.Printf("  Applying changes with persistent configuration\n")
				if err := wireguard.UpdatePersistentConfig(client, m.InterfaceName, config, routes, diff); err != nil {
					return fmt.Errorf("failed to update config on %s: %w", hostname, err)
				}
			} else {
				fmt.Printf("  No changes needed, ensuring persistent config is up to date\n")
				configContent := wireguard.GenerateWgQuickConfig(config, routes)
				configPath := fmt.Sprintf("/etc/wireguard/%s.conf", m.InterfaceName)
				if err := client.WriteFile(configPath, []byte(configContent), 0600); err != nil {
					fmt.Printf("  Warning: failed to update config file: %v\n", err)
				}
			}
		}

		fmt.Printf("  ✓ Deployed successfully\n\n")
	}

	return nil
}

func (m *Mesh) detectEndpoints() error {
	for hostname, node := range m.Nodes {
		if node.IsLocal {
			continue
		}

		client, err := ssh.NewClient(node.SSHHost, node.SSHPort)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", hostname, err)
		}

		publicIP, err := ssh.DetectPublicIP(client)
		client.Close()

		if err != nil {
			fmt.Printf("Warning: failed to detect public IP for %s: %v\n", hostname, err)
			node.BehindNAT = true
			continue
		}

		if publicIP != "" && publicIP != node.SSHHost {
			node.BehindNAT = true
			fmt.Printf("Detected %s is behind NAT (public IP: %s)\n", hostname, publicIP)
		} else {
			node.PublicEndpoint = fmt.Sprintf("%s:%d", node.SSHHost, node.ListenPort)
			fmt.Printf("Detected %s has public endpoint: %s\n", hostname, node.PublicEndpoint)
		}
	}

	return nil
}

func (m *Mesh) collectRoutesForNode(node *Node) []ssh.RouteEntry {
	routes := make([]ssh.RouteEntry, 0)

	for peerHostname, peer := range m.Nodes {
		if peerHostname == node.Hostname {
			continue
		}

		for _, network := range peer.RoutableNetworks {
			routes = append(routes, ssh.RouteEntry{
				Network: network,
				Gateway: peer.MeshIP.String(),
			})
		}
	}

	return routes
}

func (m *Mesh) generateConfigForNode(node *Node) *WireGuardConfig {
	config := &WireGuardConfig{
		Interface: WGInterface{
			PrivateKey: node.PrivateKey,
			Address:    fmt.Sprintf("%s/16", node.MeshIP.String()),
			ListenPort: node.ListenPort,
		},
		Peers: make([]WGPeer, 0),
	}

	for peerHostname, peer := range m.Nodes {
		if peerHostname == node.Hostname {
			continue
		}

		allowedIPs := []string{fmt.Sprintf("%s/32", peer.MeshIP.String())}

		for _, network := range peer.RoutableNetworks {
			allowedIPs = append(allowedIPs, network)
		}

		peerConfig := WGPeer{
			PublicKey:  peer.PublicKey,
			AllowedIPs: allowedIPs,
		}

		if peer.PublicEndpoint != "" {
			peerConfig.Endpoint = peer.PublicEndpoint
		}

		peerConfig.PersistentKeepalive = 5

		config.Peers = append(config.Peers, peerConfig)
	}

	return config
}
