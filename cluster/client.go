package cluster

import (
	"fmt"
	"net/rpc"
)

func (n *Node) HandleClientCommand(cmd string, key string, value interface{}){
    n.mu.Lock()
    if n.Role != "Leader" {
        n.mu.Unlock()
        fmt.Println("no leader currently known, try again later")
        return
    }
    
    entry := LogEntry{
        Term: n.CurrentTerm,
        Command: cmd,
        Key: key,
        Value: value,
    }
    n.Log = append(n.Log, entry)
    n.ackedLength[n.Address] = len(n.Log)
    n.mu.Unlock()
    
    for _, peer := range n.Peers {
        go n.ReplicateLog(peer)
    }
}

func (n *Node) HandleClientCommandRPC(args *ClientCommandArgs, reply *ClientCommandReply) error {
    n.mu.Lock()
    defer n.mu.Unlock()
    
    if n.Role != "Leader" {
        reply.Success = false
        return fmt.Errorf("not leader")
    }
    
    entry := LogEntry{
        Term:    n.CurrentTerm,
        Command: args.Command,
        Key:     args.Key,
        Value:   args.Value,
    }
    
    n.Log = append(n.Log, entry)
    n.ackedLength[n.Address] = len(n.Log)
    
    reply.Success = true
    
    // Trigger replication
    go func() {
        for _, peer := range n.Peers {
            n.ReplicateLog(peer)
        }
    }()
    
    return nil
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

