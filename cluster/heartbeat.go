package cluster

import (
	"fmt"
	"net/rpc"
	"time"
)

type AppendEntriesArgs struct {
	Term     int
	LeaderID int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

const (
	HeartbeatInterval = 50 * time.Millisecond
)


func (n *Node) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Reject if leader's term is older
	if args.Term < n.CurrentTerm {
		reply.Term = n.CurrentTerm
		reply.Success = false
		return nil
	}

	// Accept heartbeat
	n.CurrentTerm = args.Term
	n.Role = "Follower"
	n.CurrentLeader = args.LeaderID
	n.VotedFor = -1

	// Reset election timer
	n.resetElectionTimer()

	reply.Term = n.CurrentTerm
	reply.Success = true
		fmt.Printf("[Node %d] heartbeat from leader %d\n", n.ID, args.LeaderID) 

	return nil
}


func (n *Node) sendHeartbeat(peer string, args *AppendEntriesArgs) {
	client, err := rpc.Dial("tcp", peer)
	if err != nil {
		return
	}
	defer client.Close()

	reply := &AppendEntriesReply{}
	client.Call("Node.AppendEntries", args, reply)

	// If peer has higher term, step down
	if reply.Term > n.CurrentTerm {
		n.mu.Lock()
		n.Role = "Follower"
		n.CurrentTerm = reply.Term
		n.mu.Unlock()
	}
}

func (n *Node) startHeartbeat() {
	n.heartbeatTicker = time.NewTicker(HeartbeatInterval)

	go func() {
		for range n.heartbeatTicker.C {
			n.mu.Lock()
			if n.Role != "Leader" {
				n.mu.Unlock()
				return
			}
			term := n.CurrentTerm
			n.mu.Unlock()

			args := &AppendEntriesArgs{
				Term:     term,
				LeaderID: n.ID,
			}

			for _, peer := range n.Peers {
				go n.sendHeartbeat(peer, args)
			}
		}
	}()
}
