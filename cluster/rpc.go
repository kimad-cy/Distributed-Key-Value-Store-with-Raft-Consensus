package cluster

func sendRquestVote(n *Node) (cID int, cTerm int, cLogLength int, cLogTerm int){
	if cTerm > n.CurrentTerm{
		n.CurrentTerm = cTerm
		n.Role = "Follower"
		n.VotedFor = -1
	}
	lastLogTerm := 0
	if len(n.Log) > 0{
		lastLogTerm = n.Log[len(n.Log)-1].Term
	}
	ok := (cLogTerm > lastLogTerm) || (cLogTerm == lastLogTerm && cLogLength >= len(n.Log))

	if cTerm == n.CurrentTerm && ok && (n.VotedFor == cID || n.VotedFor == -1){
		n.VotesReceived = c.ID
		sendVoteResponse()
	}

}

