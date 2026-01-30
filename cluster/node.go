package cluster

import (
	"sync"
	"time"
	"Distributed-Key-Value-Store-with-Raft-Consensus/store"
)

type Node struct {
	//Node identity & cluster info
	ID int `json:"id"`
	Address string `json:"address"` 
	Peers []string `json:"peers"`

	//Raft role & term state 
	Role string `json:"role"`
	CurrentTerm int
	VotedFor int 
	CurrentLeader int 

	//Raft Log
	Log []LogEntry `json:"log"`

	//Commit & Apply State
	CommitIdx int `json:"commit_index"`
	lastApplied int
	
	//Leader Election
	VotesReceived map[string]bool
	ElectionTimer *time.Timer

	//Heartbeats
	heartbeatTicker *time.Ticker

	// Leader-only 
    sentLength  map[string]int   // followerAddr → last sent index
    ackedLength map[string]int   // followerAddr → last acked index

	//State Machine
	Store     *store.KVStore 

	PeerIDToAddr map[int]string  // Maps node ID to address

	//Concurrency
	mu sync.RWMutex
}


//Constructor
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
		sentLength: make(map[string]int), 
        ackedLength: make(map[string]int),
		PeerIDToAddr:  make(map[int]string),
	}
	return &node
}

func (n *Node) GetID() int {
    return n.ID
}

func (n *Node) GetRole() string {
    n.mu.RLock()
    defer n.mu.RUnlock()
    return n.Role
}

func (n *Node) GetCurrentTerm() int {
    n.mu.RLock()
    defer n.mu.RUnlock()
    return n.CurrentTerm
}

func (n *Node) GetKVSnapshot() map[string]interface{} {
    return n.Store.Snapshot() 
}

func (n *Node) GetLog() []LogEntry {
    n.mu.RLock()
    defer n.mu.RUnlock()
    logCopy := make([]LogEntry, len(n.Log))
    copy(logCopy, n.Log)
    return logCopy
}
