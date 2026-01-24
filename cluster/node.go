package cluster

import (
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
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

type LogEntry struct {
	Term int `json:"term"`
	Command string `json:"command"` 
	Key string `json:"key"`
	Value interface{} `json:"value"`
}

//Construuctor
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

	for _, peer := range n.Peers {
		go func(p string) {
			n.ackedLength[p] = 0
			n.sentLength[p] = len(n.Log)
			n.ReplicateLog(n.ID, p)
		}(peer)
	}

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

		fmt.Println("Waiting to be applied")


		prevTerm := 0
		if len(n.Log) != 0{
			prevTerm = n.Log[len(n.Log)-1].Term
		}
		entries := []LogEntry{}
		entries = append(entries, entry)
		args := &AppendEntriesArgs{
			Term: n.CurrentTerm,
			LeaderID: n.ID,
			PrevLogIndex: len(n.Log),
			PrevLogTerm: prevTerm,
			Entries: entries,
			LeaderCommit: n.CommitIdx,
		}

		reply := AppendEntriesReply{}

		n.AppendEntries(args, &reply)

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

