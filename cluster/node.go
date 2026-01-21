package cluster

import (
	"fmt"
	"math/rand"
	"net/rpc"
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
	heartbeatTicker *time.Ticker
	IDToAddr map[int]string
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

	if n.ElectionTimer != nil {
		n.ElectionTimer.Stop()
	}

	fmt.Printf("[Node %d] BECAME LEADER (term %d)\n", n.ID, n.CurrentTerm)

	n.startHeartbeat()
}

/*************************  Election Timer Functions  ********************************/

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


func (n *Node) resetElectionTimer() {
	if n.ElectionTimer == nil {
		n.startElectionTimer()
		return
	}
	n.ElectionTimer.Stop()
	n.ElectionTimer.Reset(randomElectionTimeout())
}


/*************************  Handling Client Requests ********************************/

type ClientCommandArgs struct {
    Command string
    Key     string
    Value   interface{}
}

type ClientCommandReply struct {
    Success bool
}



func (n *Node) HandleClientCommand(cmd string, key string, value interface{}){
	if n.Role == "Leader" {
		entry := LogEntry{
			Term: n.CurrentTerm,
			Command: cmd,
			Key: key,
			Value: value,
		}
		n.mu.Lock()
		defer n.mu.Unlock()
		n.Log = append(n.Log, entry)

		return
	}

	n.ForwardToLeader(cmd, key, value)
}

func (n *Node) ForwardToLeader(cmd string, key string, value interface{}) error {
    n.mu.RLock()
    leaderID := n.CurrentLeader
    n.mu.RUnlock()

    if leaderID == -1 {
        return fmt.Errorf("no leader currently known, try again later")
    }

    var leaderAddr string
    // for _, addr := range n.Peers {
    //     // find the address of the leader 
    // }
    if leaderAddr == "" {
        return fmt.Errorf("leader address not found")
    }

    // Prepare a request object
    args := &ClientCommandArgs{
        Command: cmd,
        Key:     key,
        Value:   value,
    }
    reply := &ClientCommandReply{}

    // Call leader via RPC
    client, err := rpc.Dial("tcp", leaderAddr)
    if err != nil {
        return fmt.Errorf("failed to connect to leader: %v", err)
    }
    defer client.Close()

    err = client.Call("Node.HandleClientCommandRPC", args, reply)
    if err != nil {
        return fmt.Errorf("leader RPC failed: %v", err)
    }

    if !reply.Success {
        return fmt.Errorf("leader rejected command")
    }

    return nil
}

