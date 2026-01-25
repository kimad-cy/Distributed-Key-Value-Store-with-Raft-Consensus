package cluster

import (
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
				go n.ReplicateLog(peer)
			}
		}
	}()
}
