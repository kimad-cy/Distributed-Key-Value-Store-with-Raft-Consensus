package cluster

import (
	//"fmt"
	"time"
)


const (
	HeartbeatInterval = 50 * time.Millisecond
)

func (n *Node) StartHeartbeat() {
	n.heartbeatTicker = time.NewTicker(HeartbeatInterval)

	go func() {
		for range n.heartbeatTicker.C {
			n.mu.Lock()
			if n.Role != "Leader" {
				if n.heartbeatTicker != nil{
					n.heartbeatTicker.Stop()
				}
				n.mu.Unlock()
				return
			}
			
			n.mu.Unlock()

			for _, peer := range n.Peers {
				//fmt.Printf("[Node %d] sent heartbeat to %s\n", n.ID, peer)
				go n.ReplicateLog(peer)
			}
		}
	}()
}
