package cluster

import (
	"fmt"
	"math/rand"
	"time"
)

/************************ Election Timers **************************************/

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

/**************************************************************/

func (n *Node) StartElection() {
	n.mu.Lock()
	n.CurrentTerm++
	n.Role = "Candidate"
	n.VotedFor = n.ID
	n.VotesReceived = make([]int, 1)
	n.mu.Unlock()

	fmt.Printf("[Node %d] starting election (term %d)\n", n.ID, n.CurrentTerm)

	lastLogTerm := 0
	if len(n.Log) > 0{
		lastLogTerm = n.Log[len(n.Log)-1].Term
	}

	args := &RequestVoteArgs{
		Term:         n.CurrentTerm,
		CandidateID:  n.ID,
		LastLogIndex: len(n.Log),
		LastLogTerm:  lastLogTerm,
	}

	for _, peer := range n.Peers {
		go func(p string) {
			reply, err := n.sendRequestVote(p, args)
			if err != nil {
				return
			}
			n.processVoteReply(reply)
		}(peer)
	}
}

/*************************** Votes Functions ***********************************/

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


func (n *Node) processVoteReply(reply *RequestVoteReply) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// If reply has higher term: step down
	if reply.Term > n.CurrentTerm {
		n.CurrentTerm = reply.Term
		n.Role = "Follower"
		n.VotedFor = -1
		n.CurrentLeader = -1
		n.VotesReceived = nil
		return
	}

	// Ignore replies if no longer candidate
	if n.Role != "Candidate" {
		return
	}

	// Count granted votes
	if reply.VoteGranted {
		n.VotesReceived = append(n.VotesReceived, 1) 
		fmt.Printf("[Node %d] received vote (%d total)\n", n.ID, len(n.VotesReceived))

		// Majority check
		if len(n.VotesReceived) >= (len(n.Peers)+1)/2+1 {
			n.becomeLeader()
		}
	}
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
			n.ReplicateLog(p)
		}(peer)
	}

	n.StartHeartbeat()
}