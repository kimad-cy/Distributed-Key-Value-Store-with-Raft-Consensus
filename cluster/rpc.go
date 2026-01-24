package cluster

import (
	"fmt"
	"net"
	"net/rpc"
)


type RequestVoteArgs struct {
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

func (n *Node) sendRequestVote(peerAddr string,args *RequestVoteArgs,) (*RequestVoteReply, error) {

	client, err := rpc.Dial("tcp", peerAddr)
	if err != nil {
		fmt.Printf("[Node %d] ERROR connecting to %s: %v\n", n.ID, peerAddr, err)
		return nil, err
	}
	defer client.Close()

	reply := &RequestVoteReply{}
	err = client.Call("Node.RequestVote", args, reply)
	return reply, err
}


func (n *Node) RequestVote(args *RequestVoteArgs,reply *RequestVoteReply,) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	fmt.Printf("[Node %d] received RequestVote from %d (term %d)\n",
		n.ID, args.CandidateID, args.Term)

	// Reply false if term is older
	if args.Term < n.CurrentTerm {
		reply.Term = n.CurrentTerm
		reply.VoteGranted = false
		return nil
	}

	// If term is newer, update self
	if args.Term > n.CurrentTerm {
		n.CurrentTerm = args.Term
		n.Role = "Follower"
		n.VotedFor = -1
		n.CurrentLeader = -1 
	}

	// Check log freshness
	lastLogTerm := 0
	if len(n.Log) > 0 {
		lastLogTerm = n.Log[len(n.Log)-1].Term
	}

	upToDate := args.LastLogTerm > lastLogTerm ||
		(args.LastLogTerm == lastLogTerm && args.LastLogIndex >= len(n.Log))

	// Grant vote
	if (n.VotedFor == -1 || n.VotedFor == args.CandidateID) && upToDate {
		n.VotedFor = args.CandidateID
		reply.VoteGranted = true
		fmt.Printf("[Node %d] voted for %d\n", n.ID, args.CandidateID)
	} else {
		reply.VoteGranted = false
	}

	reply.Term = n.CurrentTerm
	return nil
}

func (n *Node) Start() error {
	rpc.Register(n)

	listener, err := net.Listen("tcp", n.Address)
	if err != nil {
		return err
	}

	fmt.Printf("[Node %d] listening on %s\n", n.ID, n.Address)
	go rpc.Accept(listener)

	n.startElectionTimer()
	
	return nil
}



