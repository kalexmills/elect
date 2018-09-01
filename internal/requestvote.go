package elect


// RequestVoteQ is the request for RequestVote messages.
type RequestVoteQ struct {
	term uint64
	candidateId uint64
	lastLogIndex uint64
	lastLogTerm uint64
}

// RequestVoteA is the response to RequestVote messages.
type RequestVoteA struct {
	term uint64
	voteGranted bool
}

// RequestVote is the RPC handler that handles RequestVote calls
func (node *Node) RequestVote(in RequestVoteQ, out *RequestVoteA) error {
	state := node.localstate
	state.OnReceiveRPC(in.term)

	out.term = state.currentTerm
	if in.term < state.currentTerm {
		out.voteGranted = false
	}
	state.votedFor = state.tryVote(in.candidateId)
	if state.votedFor == in.candidateId {
		// TODO: Check that the candidate's log is at least as up-to-date as my own.
		out.voteGranted = true
	}
	return nil
}

