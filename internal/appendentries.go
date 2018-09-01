package elect

// AppendEntriesQ is the request to AppendEntries messages.
type AppendEntriesQ struct {
	Term         uint64
	LeaderId     uint64
	PrevLogIndex uint64
	PrevLogTerm  uint64
	Entries      []int
	LeaderCommit uint64
}

// AppendEntriesA is the response to AppendEntries messages.
type AppendEntriesA struct {
	Term    uint64
	Success bool
}

// AppendEntries is the RPC Handler that handles AppendEntries calls
func (node *Node) AppendEntries(in AppendEntriesQ, out *AppendEntriesA) error {
	state := node.localstate
	state.OnReceiveRPC(in.Term)

	state.ResetTimer()
	state.SignalCandidateLost()
	state.logger.Println("AppendEntries message received during Term ", in.Term)

	// For leader-election, all that is needed is that this serves as a heart-beat.
	return nil
}

