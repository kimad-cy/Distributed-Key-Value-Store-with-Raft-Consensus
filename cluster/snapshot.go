package cluster

import (
	"encoding/json"
	"fmt"
	"os"
)

func (n *Node) TakeSnapshot() {
    n.mu.Lock()
    defer n.mu.Unlock()

    // Only snapshot if we have applied entries
    if n.CommitIdx <= n.LastIncludedIndex {
        return
    }

    snapshotData := n.Store.Snapshot()
    
    n.LastIncludedIndex = n.CommitIdx
    n.LastIncludedTerm = n.Log[n.CommitIdx-1].Term
    
    // Log index 0 now represents LastIncludedIndex + 1
    newLog := make([]LogEntry, len(n.Log)-n.CommitIdx)
    copy(newLog, n.Log[n.CommitIdx:])
    n.Log = newLog

    // Save to disk
    n.persistSnapshot(snapshotData)
    fmt.Printf("[Node %d] Snapshot taken up to index %d\n", n.ID, n.LastIncludedIndex)
}

func (n *Node) persistSnapshot(data map[string]interface{}) {
    filename := fmt.Sprintf("snapshot-%d.json", n.ID)
    file, err := os.Create(filename)
    if err != nil {
        return 
    }
    defer file.Close()
    
    json.NewEncoder(file).Encode(data)
}

func (n *Node) readSnapshot() {
    filename := fmt.Sprintf("snapshot-%d.json", n.ID)
    file, err := os.Open(filename)
    if err != nil {
        return
    }
    defer file.Close()

    var data map[string]interface{}
    if err := json.NewDecoder(file).Decode(&data); err != nil {
        fmt.Printf("[Node %d] Error decoding snapshot: %v\n", n.ID, err)
        return
    }

    // Restore the KVStore with this data
    n.Store.Restore(data)
    
    n.CommitIdx = n.LastIncludedIndex
    n.lastApplied = n.LastIncludedIndex
    
    fmt.Printf("[Node %d] Loaded snapshot with %d keys\n", n.ID, len(data))
}