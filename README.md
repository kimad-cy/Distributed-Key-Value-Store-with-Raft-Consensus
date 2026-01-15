# Distributed-Key-Value-Store-with-Raft-Consensus

1- Core Node State & Local Storage
  ~ Implement a thread-safe in-memory key-value store.
  
  ~ Apply committed log entries to local state.
  
  ~ Ensure all shared state is protected with locks.

2- Client REST API

  ~ Expose GET, SET, and cluster status endpoints.

  ~ Accept writes only on the leader.

  ~ Forward requests to the leader when necessary.

3- RPC Communication Layer

  ~ Enable node-to-node communication using RPC.

  ~ Handle timeouts and unreachable peers gracefully.

  ~ Support concurrent requests between nodes.

4-  Node Roles & State Machine

  ~ Manage transitions between Follower, Candidate, and Leader.

  ~ Track terms and leader identity consistently.

  ~ Reset timers correctly on role changes.

5- Leader Election

  ~ Trigger elections on heartbeat timeouts.

  ~ Collect votes concurrently from peers.

  ~ Elect a leader only with majority agreement.

6- Heartbeat Mechanism

  ~ Send periodic heartbeats from the leader.

  ~ Prevent unnecessary elections in healthy clusters.

  ~ Detect leader failure quickly.

 7- Log Replication

  ~ Replicate log entries from leader to followers.

  ~ Commit entries only after majority confirmation.

  ~ Ensure logs remain consistent across nodes.

8- Persistence & Recovery

  ~ Persist logs and metadata to disk.

  ~ Reload state on node restart.

  ~ Resume operation without data loss.

 9- Log Compaction

  ~ Periodically snapshot committed state.

  ~ Remove old log entries safely.

  ~ Reduce memory usage over time.

10- Cluster Status & Monitoring

  ~ Expose current leader and node roles.

  ~ Report peer health and commit progress.

  ~ Aid debugging and evaluation.
