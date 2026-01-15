package api

import (
	"Distributed-Key-Value-Store-with-Raft-Consensus/cluster"
)

func NewRouter(node *cluster.Node) *mux.Router {
    router := mux.NewRouter()
    handlers := &Handlers{Node: node}
    
    router.HandleFunc("/get/{key}", handlers.GetKey).Methods("GET")
    router.HandleFunc("/set", handlers.SetKey).Methods("POST")
    router.HandleFunc("/cluster/status", handlers.ClusterStatus).Methods("GET")
    
    return router
}