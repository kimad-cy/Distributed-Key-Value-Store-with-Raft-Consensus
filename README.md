# Distributed KV Store – TODO Checklist

> High-level implementation checklist 

---

## ⬜ 1. Core Node State & Local Storage

* ⬜ Initialize Node with ID, Address, Peers, Role, Log, CommitIdx
* ⬜ Create thread-safe in-memory key-value store
* ⬜ Apply committed log entries to local state
* ⬜ Ensure all shared state is protected with mutexes

---

## ⬜ 2. Client REST API

* ⬜ Implement GET /get/{key}
* ⬜ Implement POST /set
* ⬜ Implement GET /cluster/status
* ⬜ Accept writes only on leader
* ⬜ Forward client requests to leader when needed

---

## ⬜ 3. RPC Communication Layer

* ⬜ Set up RPC server on each node
* ⬜ Enable node-to-node RPC clients
* ⬜ Implement timeouts using context.WithTimeout
* ⬜ Handle unreachable or slow peers safely

---

## ⬜ 4. Node Roles & State Machine

* ⬜ Support Follower, Candidate, Leader roles
* ⬜ Track current term and leader identity
* ⬜ Implement safe role transitions
* ⬜ Reset timers on role or term changes

---

## ⬜ 5. Leader Election

* ⬜ Trigger elections on heartbeat timeout
* ⬜ Start election as Candidate
* ⬜ Send RequestVote RPCs concurrently
* ⬜ Count votes and detect majority
* ⬜ Become Leader only on majority
* ⬜ Step down on higher term detection

---

## ⬜ 6. Heartbeat Mechanism

* ⬜ Start periodic heartbeats when Leader
* ⬜ Send empty AppendEntries to peers
* ⬜ Reset follower election timers on heartbeat
* ⬜ Detect leader failure via missed heartbeats

---

## ⬜ 7. Log Replication

* ⬜ Append new entries to leader log
* ⬜ Replicate entries to peers concurrently
* ⬜ Commit entries after majority replication
* ⬜ Apply committed entries in order
* ⬜ Handle offline or slow nodes gracefully

---

## ⬜ 8. Persistence & Recovery

* ⬜ Persist log and commit index to disk
* ⬜ Persist current term
* ⬜ Load persisted state on startup
* ⬜ Resume operation after crash

---

## ⬜ 9. Log Compaction

* ⬜ Run periodic background compaction task
* ⬜ Snapshot current key-value state
* ⬜ Remove committed log entries safely
* ⬜ Prevent unbounded log growth

---

## ⬜ 10. Cluster Status & Monitoring

* ⬜ Report current node role
* ⬜ Report leader identity
* ⬜ Report peer health
* ⬜ Expose replication and commit progress

---

## ✅ Completion Criteria

* ⬜ Data replicated to at least 2/3 nodes
* ⬜ Leader re-elected correctly on failure
* ⬜ No race conditions under concurrent SETs
* ⬜ Nodes recover correctly after restart
* ⬜ 3-node cluster can be started and documented

---

> This checklist defines a **minimum complete and gradable implementation**.
> Extra optimizations are optional and not required for full score.
