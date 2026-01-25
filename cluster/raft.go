package cluster

import (
	"Distributed-Key-Value-Store-with-Raft-Consensus/store"
	"context"
	"fmt"
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

	n.resetElectionTimer()
	
	if args.Term > n.CurrentTerm{
		n.CurrentTerm = args.Term
		n.VotedFor = -1
	}
	

    n.CurrentLeader = args.LeaderID
    n.Role = "Follower"
	
	if len(args.Entries) == 0 {
		// Heartbeat
		n.resetElectionTimer()   
		reply.Success = true
		reply.Term = n.CurrentTerm
		if n.ID != args.LeaderID { 
			fmt.Printf("[Node %d] Received heartbeat from %d\n", n.ID, args.LeaderID)
		}
		return nil
	}

	if (args.PrevLogIndex <= len(n.Log)) && (args.PrevLogIndex == 0 || args.PrevLogTerm == n.Log[args.PrevLogIndex-1].Term){
		// Reset election timer
		n.resetElectionTimer()

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
			for i := 0; i < len(args.Entries); i++ {
				logIdx := args.PrevLogIndex + i
				if logIdx < len(n.Log) {
					if n.Log[logIdx].Term != args.Entries[i].Term {
						n.Log = n.Log[:logIdx]
						break
					}
				}
			}
		}else{
			n.Log = n.Log[0:args.PrevLogIndex]
		}
		
	}

	if args.PrevLogIndex+ len(args.Entries) > len(n.Log) {
		start:= max(0,len(n.Log) - args.PrevLogIndex)
		for i:=start; i< len(args.Entries); i++{
			n.Log = append(n.Log, args.Entries[i])
		}

	}

	if args.LeaderCommit > n.CommitIdx {
		for n.CommitIdx < args.LeaderCommit {
        raftEntry := n.Log[n.CommitIdx]

        storeEntry := store.LogEntry{
            Term: raftEntry.Term,
            Command: raftEntry.Command,
            Key: raftEntry.Key,
            Value: raftEntry.Value,
        }

        n.Store.Apply(storeEntry)
        n.CommitIdx++
    }
	}

}

/*****************************Leader Side ************************************/

func(n *Node) ReplicateLog(FollowerAddr string){
	n.mu.Lock()
    if n.Role != "Leader" {
        n.mu.Unlock()
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

	n.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), RPCTimeout)
    defer cancel()

    // Dial with context
    client, err := dialWithContext(ctx, "tcp", FollowerAddr)
    if err != nil {
        fmt.Printf("[Node %d] ERROR connecting to %s: %v\n", n.ID, FollowerAddr, err)
        return
    }
    defer client.Close()

    // Make RPC call with context
    reply := &AppendEntriesReply{}
    err = callRPCWithContext(ctx, client, "Node.AppendEntries", args, reply)
    if err != nil {
        if err == context.DeadlineExceeded {
            fmt.Printf("[Node %d] RPC timeout to %s\n", n.ID, FollowerAddr)
        } else {
            fmt.Printf("[Node %d] RPC error to %s: %v\n", n.ID, FollowerAddr, err)
        }
        return
    }

    n.HandleAppendEntriesReply(FollowerAddr, *reply)
}

func (n *Node) HandleAppendEntriesReply(followerAddr string, resp AppendEntriesReply){
	n.mu.Lock()
    defer n.mu.Unlock()

	if resp.Term == n.CurrentTerm && n.Role == "Leader" {

		if resp.Success && resp.Ack >= n.ackedLength[followerAddr]{
			n.sentLength[followerAddr] = resp.Ack
			n.ackedLength[followerAddr] = resp.Ack
			n.CommitLogEntries()

		}else if n.sentLength[followerAddr] > 0{
			n.sentLength[followerAddr] --
			go func(addr string) {
                n.ReplicateLog(addr)
            }(followerAddr)
		}

	}else if resp.Term > n.CurrentTerm{
		n.CurrentTerm = resp.Term
		n.Role = "Follower"
		n.VotedFor = -1
		n.CurrentLeader = -1
		n.resetElectionTimer()
	}
}

func (n *Node) CommitLogEntries() {
    if n.Role != "Leader" {
        return
    }
    
    minAcks := (len(n.Peers) + 1) / 2 + 1

	//fmt.Printf("[Node %d] CommitLogEntries: checking %d log entries, CommitIdx=%d, minAcks=%d\n", 
      //  n.ID, len(n.Log), n.CommitIdx, minAcks)
    
    // Debug: print ackedLength
    //fmt.Printf("[Node %d] ackedLength: %+v\n", n.ID, n.ackedLength)
    
    for i := len(n.Log) - 1; i >= n.CommitIdx; i-- {
        acksLen := 0
        
        for _, ackedIdx := range n.ackedLength {
            if ackedIdx > i {  
                acksLen++
				//fmt.Printf("[Node %d]   Entry %d: peer %s has acked up to %d\n", 
                   // n.ID, i, addr, ackedIdx)
            }
        }

		//fmt.Printf("[Node %d]   Entry %d: acksLen=%d, term=%d, currentTerm=%d\n", 
            //n.ID, i, acksLen, n.Log[i].Term, n.CurrentTerm)
        
        // If we have majority and it's from current term, commit
        if acksLen >= minAcks && n.Log[i].Term == n.CurrentTerm {
			fmt.Printf("[Node %d] COMMITTING entries %d to %d\n", n.ID, n.CommitIdx, i)

            // Apply all entries from CommitIdx to i
            for j := n.CommitIdx; j <= i; j++ {
                raftEntry := n.Log[j]

                storeEntry := store.LogEntry{
                    Term:    raftEntry.Term,
                    Command: raftEntry.Command,
                    Key:     raftEntry.Key,
                    Value:   raftEntry.Value,
                }

                n.Store.Apply(storeEntry)
				fmt.Printf("[Node %d] Applied entry %d: %s %s=%v\n", 
                    n.ID, j, raftEntry.Command, raftEntry.Key, raftEntry.Value)
            }
            n.CommitIdx = i + 1
            break  
        }
    }
}