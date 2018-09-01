package elect

// RequestVoteQ is the request for RequestVote messages.
type RequestVoteQ struct {
	Term         uint64
	CandidateId  uint64
	LastLogIndex uint64
	LastLogTerm  uint64
}

// RequestVoteA is the response to RequestVote messages.
type RequestVoteA struct {
	SenderId    uint64
	Term        uint64
	VoteGranted bool
}

// RequestVote is the RPC handler that handles RequestVote calls
func (node *Node) RequestVote(in RequestVoteQ, out *RequestVoteA) error {
	state := node.localstate
	state.onReceiveRpc(in.Term)

	out.SenderId = state.id
	if in.Term < state.currentTerm {
		out.VoteGranted = false
	}
	out.Term = state.currentTerm
	state.votedFor = state.tryVote(in.CandidateId)

	if state.votedFor == in.CandidateId {
		// TODO: Check that the candidate's log is at least as up-to-date as my own.
		out.VoteGranted = true
	}

	if out.VoteGranted {
		state.log("RequestVote in term ", in.Term, " received from candidate ", in.CandidateId, "...Responded YES")
	} else {
		state.log("RequestVote in term ", in.Term, " received from candidate ", in.CandidateId, "...Responded NO; Voted for ", state.votedFor)
	}
	return nil
}
