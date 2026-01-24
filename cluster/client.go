package cluster

import (
	"fmt"
	"net/rpc"
)


type ClientCommandArgs struct {
    Command string
    Key     string
    Value   interface{}
}

type ClientCommandReply struct {
    Success bool
}



func (n *Node) HandleClientCommand(cmd string, key string, value interface{}){
	if n.Role == "Leader" {
		entry := LogEntry{
			Term: n.CurrentTerm,
			Command: cmd,
			Key: key,
			Value: value,
		}
		n.mu.Lock()
		defer n.mu.Unlock()
		n.Log = append(n.Log, entry)

		fmt.Println("Waiting to be applied")


		prevTerm := 0
		if len(n.Log) != 0{
			prevTerm = n.Log[len(n.Log)-1].Term
		}
		entries := []LogEntry{}
		entries = append(entries, entry)
		args := &AppendEntriesArgs{
			Term: n.CurrentTerm,
			LeaderID: n.ID,
			PrevLogIndex: len(n.Log),
			PrevLogTerm: prevTerm,
			Entries: entries,
			LeaderCommit: n.CommitIdx,
		}

		reply := AppendEntriesReply{}

		n.AppendEntries(args, &reply)

		return
	}

	n.ForwardToLeader(cmd, key, value)
}

func (n *Node) ForwardToLeader(cmd string, key string, value interface{}) error {
    n.mu.RLock()
    leaderID := n.CurrentLeader
    n.mu.RUnlock()

    if leaderID == -1 {
        return fmt.Errorf("no leader currently known, try again later")
    }

    var leaderAddr string
    // for _, addr := range n.Peers {
    //     // find the address of the leader 
    // }
    if leaderAddr == "" {
        return fmt.Errorf("leader address not found")
    }

    // Prepare a request object
    args := &ClientCommandArgs{
        Command: cmd,
        Key:     key,
        Value:   value,
    }
    reply := &ClientCommandReply{}

    // Call leader via RPC
    client, err := rpc.Dial("tcp", leaderAddr)
    if err != nil {
        return fmt.Errorf("failed to connect to leader: %v", err)
    }
    defer client.Close()

    err = client.Call("Node.HandleClientCommandRPC", args, reply)
    if err != nil {
        return fmt.Errorf("leader RPC failed: %v", err)
    }

    if !reply.Success {
        return fmt.Errorf("leader rejected command")
    }

    return nil
}

