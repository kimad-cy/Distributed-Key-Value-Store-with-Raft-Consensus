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
	VotesReceived []int
	ElectionTimer *time.Timer

	//Heartbeats
	heartbeatTicker *time.Ticker

	// Leader-only 
    sentLength  map[string]int   // followerAddr → last sent index
    ackedLength map[string]int   // followerAddr → last acked index

	//State Machine
	Store     *store.KVStore 

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
	}
	return &node
}


