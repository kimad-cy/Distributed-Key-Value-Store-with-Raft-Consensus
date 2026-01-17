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
	CurrentTerm int
    VotedFor int  
	lastApplied int
	CurrentLeader int
	VotesReceived []int
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
		Store: store.NewKVStore(),

	}
	return &node
}

func (n *Node) recoverState(){
	n.Role = "Follower"
	n.CurrentLeader = -1
}

func (n *Node) startElection(){
	n.CurrentTerm ++
	n.Role = "Candidate"
	n.VotedFor = n.ID
	n.VotesReceived = append(n.VotesReceived, n.ID)
	lastLogTerm := 0
	if len(n.Log) > 0{
		lastLogTerm = n.Log[len(n.Log)-1].Term
	}
	for _, node := n.Peers{
		node.sendRequestVote(n.ID, n.CurrentTerm, len(n.Log), lastLogTerm)
	}
	
}
