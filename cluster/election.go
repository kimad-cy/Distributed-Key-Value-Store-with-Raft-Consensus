package cluster

import (
	"fmt"
	"math/rand"
	"time"
)

/************************ Election Timers **************************************/

func randomElectionTimeout() time.Duration {
    return time.Duration(500+rand.Intn(300)) * time.Millisecond
}

func (n *Node) startElectionTimer() {
    n.resetElectionTimer()
}


func (n *Node) resetElectionTimer() {
    if n.ElectionTimer != nil {
        n.ElectionTimer.Stop()
    }
    
    n.ElectionTimer = time.AfterFunc(randomElectionTimeout(), func() {
        n.mu.Lock()
        // If already leader, do nothing
        if n.Role == "Leader" {
            n.mu.Unlock()
            return
        }
        n.mu.Unlock()
        
        fmt.Printf("[Node %d] election timeout\n", n.ID)
        go n.StartElection()
    })
}


/**************************************************************/

func (n *Node) StartElection() {
	n.mu.Lock()
	n.CurrentTerm++
	n.Role = "Candidate"
	n.VotedFor = n.ID
	n.VotesReceived = make(map[string]bool)
	n.VotesReceived[n.Address] = true

	lastLogTerm := 0
	if len(n.Log) > 0{
		lastLogTerm = n.Log[len(n.Log)-1].Term
	}

	currentTerm := n.CurrentTerm 
	myAddress := n.Address  
    peers := n.Peers
	n.Persist()

	n.mu.Unlock()

	fmt.Printf("[Node %d] starting election (term %d), sending to peers: %v\n", n.ID, currentTerm, peers)
	n.resetElectionTimer()

	args := &RequestVoteArgs{
		Term:         currentTerm,
		CandidateID:  n.ID,
		LastLogIndex: len(n.Log),
		LastLogTerm:  lastLogTerm,
	}

	for _, peer := range n.Peers {
		if peer == myAddress {
			fmt.Print("Shouldnt request to self!!!!!!")
            continue  // Skip self
        }
		go func(p string) {
			fmt.Printf("[Node %d] Sending RequestVote to %s\n", n.ID, p)
			reply, err := n.sendRequestVote(p, args)
			if err != nil {
				return
			}
			fmt.Printf("[Node %d] Received vote reply from %s: granted=%v\n", n.ID, p, reply.VoteGranted)
			n.processVoteReply(p,reply)
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

	// If we're in the same term and we know there's a valid leader, reject
	if args.Term == n.CurrentTerm && n.CurrentLeader != -1 {
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
		n.resetElectionTimer()
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
		n.resetElectionTimer()  
		fmt.Printf("[Node %d] voted for %d\n", n.ID, args.CandidateID)
	} else {
		reply.VoteGranted = false
	}

	reply.Term = n.CurrentTerm
	n.Persist()

	return nil
}


func (n *Node) processVoteReply(peer string, reply *RequestVoteReply) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// If reply has higher term: step down
	if reply.Term > n.CurrentTerm {
		n.CurrentTerm = reply.Term
		n.Role = "Follower"
		n.VotedFor = -1
		n.CurrentLeader = -1
		n.VotesReceived = nil
		n.resetElectionTimer()
		return
	}

	// Ignore replies if no longer candidate
	if n.Role != "Candidate" {
		return
	}

	// Count granted votes
	if reply.VoteGranted {
		n.VotesReceived[peer] = true

		// Majority check
		voterVotes := 0
		for voter := range n.VotesReceived {
			if n.VotesReceived[voter] {
				voterVotes++
			}
		}
		fmt.Printf("[Node %d] received vote (%d total)\n", n.ID, voterVotes)

		majority := (len(n.Peers) + 1 + 1) / 2
		if voterVotes >= majority{
			n.becomeLeader()
		}

	}
	n.Persist()

}

func (n *Node) becomeLeader() {
	n.Role = "Leader"
	n.CurrentLeader = n.ID

	for _, peer := range n.Peers {
        n.ackedLength[peer] = 0
        n.sentLength[peer] = len(n.Log)
    }

	n.ackedLength[n.Address] = len(n.Log)
    n.sentLength[n.Address] = len(n.Log)

	if n.ElectionTimer != nil {
		n.ElectionTimer.Stop()
	}

	fmt.Printf("[Node %d] BECAME LEADER (term %d)\n", n.ID, n.CurrentTerm)

	for _, peer := range n.Peers {
		go n.ReplicateLog(peer)
	}

	if n.heartbeatTicker != nil {
		n.heartbeatTicker.Stop()
	}
	n.StartHeartbeat()
}