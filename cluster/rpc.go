package cluster

import (
	"fmt"
	"net"
	"net/rpc"
)

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



