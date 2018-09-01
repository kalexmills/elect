package elect

// AppendEntriesQ is the request to AppendEntries messages.
type AppendEntriesQ struct {
	term uint64
	leaderId uint64
	prevLogIndex uint64
	prevLogTerm uint64
	entries []int
	leaderCommit uint64
}

// AppendEntriesA is the response to AppendEntries messages.
type AppendEntriesA struct {
	term uint64
	success bool
}

// AppendEntries is the RPC Handler that handles AppendEntries calls
func (node *Node) AppendEntries(in AppendEntriesQ, out *AppendEntriesA) error {
	state := node.localstate
	state.OnReceiveRPC(in.term)

	// For leader-election, all that is needed is that this serves as a heart-beat.
	return nil
}

