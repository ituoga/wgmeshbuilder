package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ituoga/wgmeshbuilder/pkg/mesh"
)

func main() {
	var (
		stateFile  = flag.String("state", "mesh-state.json", "Path to mesh state file")
		addNode    = flag.String("add", "", "Add node (format: hostname:ip:ssh_host[:ssh_port])")
		removeNode = flag.String("remove", "", "Remove node by hostname")
		list       = flag.Bool("list", false, "List all nodes")
		deploy     = flag.Bool("deploy", false, "Deploy configuration to all nodes")
		init       = flag.Bool("init", false, "Initialize new mesh")
	)

	flag.Parse()

	if *init {
		if err := mesh.Initialize(*stateFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize mesh: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Mesh initialized successfully")
		return
	}

	m, err := mesh.Load(*stateFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load mesh state: %v\n", err)
		os.Exit(1)
	}

	switch {
	case *addNode != "":
		if err := m.AddNode(*addNode); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add node: %v\n", err)
			os.Exit(1)
		}
		if err := m.Save(*stateFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save state: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Node added successfully\n")

	case *removeNode != "":
		if err := m.RemoveNode(*removeNode); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove node: %v\n", err)
			os.Exit(1)
		}
		if err := m.Save(*stateFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save state: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Node removed successfully\n")

	case *list:
		m.List()

	case *deploy:
		if err := m.Deploy(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to deploy: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Deployment completed successfully")

	default:
		flag.Usage()
		os.Exit(1)
	}
}
