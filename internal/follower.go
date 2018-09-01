package elect

// Follower runs the protocol for a follower node. Returns the next node type it would like to transition to.
func (state *State) Follower() (newState int) {
	state.resetElectionTimer()
	state.votedFor = Noone

	// Followers are boring... they basically just respond to RPCs
	for {
		select {
		case <-state.electionTimeout:
			state.log("Election timeout occurred")
			return Candidate
		}
	}
}
