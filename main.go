package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"Distributed-Key-Value-Store-with-Raft-Consensus/api"
	"Distributed-Key-Value-Store-with-Raft-Consensus/cluster"
)

func main() {
	// Cluster configuration
	nodes := make([]*cluster.Node, 0)
	addresses := []string{
		"127.0.0.1:8001",
		"127.0.0.1:8002",
		"127.0.0.1:8003",
		"127.0.0.1:8004",
	}
	
	for i, addr := range addresses {
		peers := make([]string, 0)
		for _, p := range addresses {
			if p != addr {
				peers = append(peers, p)
			}
		}
		fmt.Printf("Node %d at %s has peers: %v\n", i+1, addr, peers)
		node := cluster.NewNode(i+1, addr, peers)

		node.ReadPersist()

		nodes = append(nodes, node)

		apiPort := fmt.Sprintf(":900%d", i+1) 
		router := api.NewRouter(node)
		
		go func(p string, h http.Handler) {
			fmt.Printf("[Node %d] API listening on %s\n", node.ID, p)
			if err := http.ListenAndServe(p, h); err != nil {
				fmt.Printf("API Server failed: %v\n", err)
			}
		}(apiPort, router)
	}
	
	for _, n := range nodes {
		for i, addr := range addresses {
			n.PeerIDToAddr[i+1] = addr // Map Node ID (1, 2, 3) to its Address
		}
	}

	// Start all nodes
	for _, n := range nodes {
		err := n.Start()
		if err != nil {
			fmt.Printf("Node %d failed to start: %v\n", n.ID, err)
		}
	}

	// // Give nodes some time to start
	// time.Sleep(1*time.Second)

	// // Wait for leader election
	// fmt.Println("Waiting for leader election...")
	// time.Sleep(3 * time.Second) // election timeout > 150-300ms

	// var leader *cluster.Node
	// for _, n := range nodes {
	// 	if n.GetRole() == "Leader" {
	// 	leader = n
	// 	break
	// }

	// }

	// if leader == nil {
	// 	fmt.Println("No leader elected!")
	// 	return
	// }

	// fmt.Printf("Leader elected: Node %d (term %d)\n", leader.ID, leader.CurrentTerm)

	// // Send some client commands to leader
	// commands := []struct {
	// 	cmd   string
	// 	key   string
	// 	value interface{}
	// }{
	// 	{"SET", "x", 10},
	// 	{"SET", "y", 20},
	// 	{"SET", "z", 30},
	// }

	// for _, c := range commands {
	// 	fmt.Printf("Sending command %s %s=%v to leader %d\n", c.cmd, c.key, c.value, leader.ID)
	// 	leader.HandleClientCommand(c.cmd, c.key, c.value)
	// 	time.Sleep(100 * time.Millisecond)
	// }

	// Wait for replication & commits
	

	// newAddr := "127.0.0.1:8004"
	// newNode := cluster.NewNode(4, newAddr, addresses)
	// newNode.CurrentLeader = leader.GetID()
	// nodes = append(nodes, newNode)

	// commands2 := []struct {
	// 	cmd   string
	// 	key   string
	// 	value interface{}
	// }{
	// 	{"SET", "x", 40},
	// 	{"SET", "w", 10},
	// }
	// for _, c := range commands2 {
	// 	fmt.Printf("Sending command %s %s=%v to leader %d\n", c.cmd, c.key, c.value, leader.ID)
	// 	leader.HandleClientCommand(c.cmd, c.key, c.value)
	// 	time.Sleep(100 * time.Millisecond)
	// }

	// // Wait for replication & commits
	// time.Sleep(2 * time.Second)
	// StateOfCluster(nodes)

	fmt.Println("Cluster is running. Press Ctrl+C to stop.")
    
    // Create a channel to receive OS signals
	stop := make(chan os.Signal, 1)
    

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop 

	fmt.Println("\nShutting down cluster...")

	StateOfCluster(nodes)
}

func StateOfCluster(nodes []*cluster.Node){
	// Check KVStore on all nodes
	fmt.Println("\nChecking KVStore on all nodes...")
	for _, n := range nodes {
		snapshot := n.Store.Snapshot()
		fmt.Printf("Node %d store: %+v\n", n.ID, snapshot)
	}

	
	// Print logs on all nodes
	fmt.Println("\nNode logs:")
	for _, n := range nodes {
		fmt.Printf("Node %d log:\n", n.ID)
		for i, entry := range n.Log {
			fmt.Printf("  %d: %+v\n", i, entry)
		}
	}
}