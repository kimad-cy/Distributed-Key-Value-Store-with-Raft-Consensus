package cluster

import (
	"net/rpc"
	"time"
)

type AppendEntriesArgs struct {
	Term     int
	LeaderID int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
	Ack int
}

const (
	HeartbeatInterval = 50 * time.Millisecond
)

/***************** Follower Side ****************/

func (n *Node) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Reject if leader's term is older
	if args.Term < n.CurrentTerm {
		reply.Term = n.CurrentTerm
		reply.Success = false
		return nil
	}
	if args.Term > n.CurrentTerm{
		n.CurrentTerm = args.Term
    	n.VotedFor = -1
		n.Role = "Follower"
		n.CurrentLeader = args.LeaderID
	}
    
	// Reset election timer
	n.resetElectionTimer()

	if (args.PrevLogIndex <= len(n.Log)) && (args.PrevLogIndex == 0 || args.PrevLogTerm == n.Log[args.PrevLogIndex-1].Term){
			n.ApplyEntries(*args)
			reply.Term = n.CurrentTerm
			reply.Success = true
			reply.Ack = args.PrevLogIndex + len(args.Entries)
			return nil
		}

	reply.Term = n.CurrentTerm
	reply.Success = false
	return nil
}

func (n *Node) ApplyEntries(args AppendEntriesArgs){
	if len(n.Log) > args.PrevLogIndex{
		if len(args.Entries) > 0 {
			index:= min(len(n.Log), args.PrevLogIndex + len(args.Entries)) - 1
			if n.Log[index].Term != args.Entries[index-args.PrevLogIndex].Term {
				n.Log = n.Log[0:args.PrevLogIndex]
			}
		}else{
			n.Log = n.Log[0:args.PrevLogIndex]
		}
		
	}

	if args.PrevLogIndex+ len(args.Entries) > len(n.Log) {
		start:= max(0,len(n.Log) - args.PrevLogIndex)
		for i:=start; i<= len(args.Entries)-1; i++{
			n.Log = append(n.Log, args.Entries[i])
		}

	}

	if args.LeaderCommit > n.CommitIdx {
		for i:= n.CommitIdx; i<= args.LeaderCommit-1; i++{
			n.Store.Apply(n.Log[i])
		}
		n.CommitIdx = args.LeaderCommit
	}

}


/***************** Leader Side *******************/


func (n *Node) HandleAppendEntriesReply(followerID int, resp AppendEntriesReply){
	if resp.Term == n.CurrentTerm && n.Role == "Leader" {
		if resp.Success && resp.Ack >= 
	}
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
