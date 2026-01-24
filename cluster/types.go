package cluster

type LogEntry struct {
	Term int `json:"term"`
	Command string `json:"command"` 
	Key string `json:"key"`
	Value interface{} `json:"value"`
}

type AppendEntriesArgs struct {
	Term     int
	LeaderID int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
	Ack int
}

type RequestVoteArgs struct {
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}
