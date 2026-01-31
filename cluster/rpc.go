package cluster

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"time"
)

func (n *Node) sendRequestVote(peerAddr string,args *RequestVoteArgs,) (*RequestVoteReply, error) {

	// Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), RPCTimeout)
    defer cancel()

    // Dial with context
    client, err := dialWithContext(ctx, "tcp", peerAddr)
    if err != nil {
        fmt.Printf("[Node %d] ERROR connecting to %s: %v\n", n.ID, peerAddr, err)
        return nil, err
    }
    defer client.Close()

    // Make RPC call with context
    reply := &RequestVoteReply{}
    err = callRPCWithContext(ctx, client, "Node.RequestVote", args, reply)
    if err != nil {
        return nil, err
    }

    return reply, nil
}


func (n *Node) Start() error {
    server := rpc.NewServer()
    err := server.Register(n)
    if err != nil {
        return err
    }

	listener, err := net.Listen("tcp", n.Address)
	if err != nil {
		return err
	}

	fmt.Printf("[Node %d] listening on %s\n", n.ID, n.Address)
	go server.Accept(listener)

	n.startElectionTimer()
	return nil
}

const RPCTimeout = 500 * time.Millisecond

func callRPCWithContext(ctx context.Context, client *rpc.Client, serviceMethod string, args interface{}, reply interface{}) error {
    call := client.Go(serviceMethod, args, reply, make(chan *rpc.Call, 1))
    
    select {
    case <-ctx.Done():
        // Context cancelled or timed out
        return ctx.Err()
    case replyCall := <-call.Done:
        // RPC completed
        return replyCall.Error
    }
}

// dialWithContext creates RPC client with connection timeout
func dialWithContext(ctx context.Context, network, address string) (*rpc.Client, error) {
    var d net.Dialer
    conn, err := d.DialContext(ctx, network, address)
    if err != nil {
        return nil, err
    }
    return rpc.NewClient(conn), nil
}


