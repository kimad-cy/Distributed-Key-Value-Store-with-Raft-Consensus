package cluster

import (
	"net/rpc"
	"time"
)


const (
	HeartbeatInterval = 50 * time.Millisecond
)



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

func (n *Node) StartHeartbeat() {
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

			prevTerm := 0
			if len(n.Log) != 0{
				prevTerm = n.Log[len(n.Log)-1].Term
			}

			args := &AppendEntriesArgs{
				Term:     term,
				LeaderID: n.ID,
				PrevLogIndex: len(n.Log),
				PrevLogTerm: prevTerm,
				LeaderCommit: n.CommitIdx,
			}

			for _, peer := range n.Peers {
				go n.sendHeartbeat(peer, args)
			}
		}
	}()
}
