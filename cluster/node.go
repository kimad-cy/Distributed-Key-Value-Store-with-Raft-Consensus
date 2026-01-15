package cluster

import (
	"sync"

	"Distributed-Key-Value-Store-with-Raft-Consensus/store"
)

type Node struct {
	ID int `json:"id"`
	Address string `json:"address"` 
	Peers []string `json:"peers"`
	Role string `json:"role"`
	Log []LogEntry `json:"log"`
	CommitIdx int `json:"commit_index"`
	mu sync.RWMutex
	Store     *store.KVStore 
}

type LogEntry struct {
	Term int `json:"term"`
	Command string `json:"command"` 
	Key string `json:"key"`
	Value interface{} `json:"value"`
}

type RaftState struct {
    CurrentTerm   int
    VotedFor      int    // ID of voted candidate
    LastLogIndex  int
    LastLogTerm   int
    StateMachine  *store.KVStore
}


func NewNode(id int, address string , peers []string) (*Node){
	node := Node{
		ID: id,
		Address: address,
		Peers: peers,
		Role: "Follower",
		Log: []LogEntry{},
		CommitIdx: 0,
		Store: store.NewKVStore(),

	}
	return &node
}


