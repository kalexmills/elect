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

	state.onReceiveRpc(in.Term)

	state.signalCandidateLost()

	// Quit early on a heartbeat.
	if len(in.Entries) == 0 {
		return nil
	}

	state.log("AppendEntries message received during Term ", in.Term)

	out.Term = state.currentTerm

	// For leader-election, all that is needed is that this serves as a heart-beat.
	return nil
}
