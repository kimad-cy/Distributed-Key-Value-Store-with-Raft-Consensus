package cluster

import "fmt"

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
