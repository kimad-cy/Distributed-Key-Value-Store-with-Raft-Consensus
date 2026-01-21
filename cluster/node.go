package cluster

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"Distributed-Key-Value-Store-with-Raft-Consensus/store"
)

type Node struct {
	ID int `json:"id"`
	Address string `json:"address"` 
	Peers []string `json:"peers"`
	Role string `json:"role"`
	Log []LogEntry `json:"log"`
	CommitIdx int `json:"commit_index"`
	CurrentTerm int
    VotedFor int  
	lastApplied int
	CurrentLeader int
	VotesReceived []int
	ElectionTimer *time.Timer
	Store     *store.KVStore 
	mu sync.RWMutex
}

type LogEntry struct {
	Term int `json:"term"`
	Command string `json:"command"` 
	Key string `json:"key"`
	Value interface{} `json:"value"`
}


func NewNode(id int, address string , peers []string) (*Node){
	node := Node{
		ID: id,
		Address: address,
		Peers: peers,
		Role: "Follower",
		Log: []LogEntry{},
		CommitIdx: 0,
		VotedFor: -1,
		CurrentLeader: -1,
		Store: store.NewKVStore(),

	}
	return &node
}

func (n *Node) becomeLeader() {
	n.Role = "Leader"
	n.CurrentLeader = n.ID
	fmt.Printf("[Node %d] BECAME LEADER (term %d)\n", n.ID, n.CurrentTerm)
}

func randomElectionTimeout() time.Duration {
    return time.Duration(150+rand.Intn(150)) * time.Millisecond
}

func (n *Node) startElectionTimer() {
	if n.ElectionTimer != nil {
        n.ElectionTimer.Stop()
    }

    timeout := randomElectionTimeout()

    
    n.ElectionTimer = time.AfterFunc(timeout, func() {
        n.mu.Lock()
		defer n.mu.Unlock()

		// If already leader, do nothing
		if n.Role == "Leader" {
			return
		}

		fmt.Printf("[Node %d] election timeout\n", n.ID)

		go n.StartElection()
    })
}


