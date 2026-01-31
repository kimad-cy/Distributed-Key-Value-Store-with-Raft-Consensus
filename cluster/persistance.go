package cluster

import (
	"encoding/json"
	"fmt"
	"os"
)

func (n *Node) Persist() {
    data := struct {
        CurrentTerm int
        VotedFor    int
        Log         []LogEntry
    }{
        CurrentTerm: n.CurrentTerm,
        VotedFor:    n.VotedFor,
        Log:         n.Log,
    }

    filename := fmt.Sprintf("node_%d_state.json", n.ID)
    file, _ := os.Create(filename)
    defer file.Close()
    
    json.NewEncoder(file).Encode(data)
}

func (n *Node) readPersist() {
    filename := fmt.Sprintf("node_%d_state.json", n.ID)
    file, err := os.Open(filename)
    if err != nil {
        return // No state saved yet
    }
    defer file.Close()

    var data struct {
        CurrentTerm int
        VotedFor    int
        Log         []LogEntry
    }
    if err := json.NewDecoder(file).Decode(&data); err == nil {
        n.CurrentTerm = data.CurrentTerm
        n.VotedFor = data.VotedFor
        n.Log = data.Log
    }
}