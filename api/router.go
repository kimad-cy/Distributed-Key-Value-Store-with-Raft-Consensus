package api

import (
    "net/http"
    "strings"
    "Distributed-Key-Value-Store-with-Raft-Consensus/cluster"

)

func NewRouter(node *cluster.Node) http.Handler {
    handlers := &Handlers{Node: node}
    
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch {
        case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/get/"):
            handlers.GetKey(w, r)
        case r.Method == "POST" && r.URL.Path == "/set":
            handlers.SetKey(w, r)
        case r.Method == "GET" && r.URL.Path == "/cluster/status":
            handlers.ClusterStatus(w, r)
        case r.URL.Path == "/ws":
            handlers.StatusWebsocket(w, r)
        case r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/node/kill"):
            handlers.Node.SimulateCrash()
            w.WriteHeader(http.StatusOK)
        default:
            http.NotFound(w, r)
        }
    })
}