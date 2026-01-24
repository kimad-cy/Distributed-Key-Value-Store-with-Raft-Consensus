package cluster

import (
	"Distributed-Key-Value-Store-with-Raft-Consensus/store"
	"fmt"
	"net/rpc"
)

/***************************** Follower Side ************************************/

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
			raftEntry := n.Log[i]

			storeEntry := store.LogEntry{
				Term: raftEntry.Term,
				Command: raftEntry.Command,
				Key:     raftEntry.Key,
				Value:   raftEntry.Value,
			}

			n.Store.Apply(storeEntry)
		}
		n.CommitIdx = args.LeaderCommit
	}

}

/*****************************Leader Side ************************************/

func(n *Node) ReplicateLog(FollowerAddr string){
	if n.Role != "Leader" {
		return
	}

	prefixLen := n.sentLength[FollowerAddr]

	if prefixLen > len(n.Log) {
		prefixLen = len(n.Log)
	}

	var prefixTerm int
	if prefixLen == 0 {
		prefixTerm = 0
	} else {
		prefixTerm = n.Log[prefixLen-1].Term
	}

	entries := n.Log[prefixLen:len(n.Log)]
	
	args := &AppendEntriesArgs{
		Term: n.CurrentTerm,
		LeaderID: n.ID,
		PrevLogIndex: prefixLen,
		PrevLogTerm: prefixTerm,
		Entries: entries,
		LeaderCommit: n.CommitIdx,
	}
	reply := &AppendEntriesReply{}

	client, err := rpc.Dial("tcp", FollowerAddr)
	if err != nil {
		fmt.Printf("[Node %d] ERROR connecting to %s: %v\n", FollowerAddr, err)
		return 
	}
	defer client.Close()

	client.Call("Node.AppendEntries", args, reply)

	n.HandleAppendEntriesReply(FollowerAddr, *reply)
}

func (n *Node) HandleAppendEntriesReply(followerAddr string, resp AppendEntriesReply){
	if resp.Term == n.CurrentTerm && n.Role == "Leader" {

		if resp.Success && resp.Ack >= n.ackedLength[followerAddr]{
			n.sentLength[followerAddr] = resp.Ack
			n.ackedLength[followerAddr] = resp.Ack
			n.CommitLogEntries()

		}else if n.sentLength[followerAddr] > 0{
			n.sentLength[followerAddr] --
			n.ReplicateLog(followerAddr)
		}

	}else if resp.Term > n.CurrentTerm{
		n.CurrentTerm = resp.Term
		n.Role = "Follower"
		n.VotedFor = -1
		n.CurrentLeader = -1
		n.resetElectionTimer()
	}
}

func (n *Node) CommitLogEntries(){
	if n.Role != "Leader"{
		return
	}
	minAcks := (len(n.Peers) + 1) / 2 + 1
	ready := []int{}
	for i:= range(n.Log){
		acksLen := 1 // Count Leader log
		for j:= range(n.ackedLength){
			if n.ackedLength[j] >= i+1 {
				acksLen ++
			}
		}
		if acksLen >= minAcks{
			ready = append(ready, i)
		}
	}
	maxReady := -1
	for i:= range(ready){
		if ready[i] > maxReady{
			maxReady = ready[i]
		}
	}
	if len(ready) > 0 && maxReady > n.CommitIdx && n.Log[maxReady -1].Term == n.CurrentTerm{
		for i:= n.CommitIdx;i<= maxReady -1; i++{
			raftEntry := n.Log[i]

			storeEntry := store.LogEntry{
				Term: raftEntry.Term,
				Command: raftEntry.Command,
				Key:     raftEntry.Key,
				Value:   raftEntry.Value,
			}

			n.Store.Apply(storeEntry)
		}
		n.CommitIdx = maxReady 
	}

}


