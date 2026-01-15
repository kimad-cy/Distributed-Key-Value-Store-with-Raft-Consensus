package api

import (
	"Distributed-Key-Value-Store-with-Raft-Consensus/cluster"
	"encoding/json"
	"net/http"
	"strings"
)

type Handlers struct {
    Node *cluster.Node
}

type SetRequest struct {
    Key   string      `json:"key"`
    Value interface{} `json:"value"`
}

type GetResponse struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value,omitempty"`
	Found bool        `json:"found"`
}

type ClusterStatusResponse struct {
	ID          int      `json:"id"`
	Role        string   `json:"role"`
	CommitIndex int      `json:"commit_index"`
	Peers       []string `json:"peers"`
}


// GET /get/{key}
func (h *Handlers) GetKey(w http.ResponseWriter, r *http.Request) {
    key := strings.TrimPrefix(r.URL.Path, "/get/")

	value, found := h.Node.Store.Get(key)

	resp := GetResponse{
		Key:   key,
		Value: value,
		Found: found,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /set
func (h *Handlers) SetKey(w http.ResponseWriter, r *http.Request) {
    var req SetRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// im consideriing the node is leader (change later)

	h.Node.Store.Set(req.Key, req.Value)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// GET /cluster/status
func (h *Handlers) ClusterStatus(w http.ResponseWriter, r *http.Request) {
    resp := ClusterStatusResponse{
		ID:          h.Node.ID,
		Role:        h.Node.Role,
		CommitIndex: h.Node.CommitIdx,
		Peers:       h.Node.Peers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

