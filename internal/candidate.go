package elect

import (
	"net/rpc"
	"strconv"
)

// Candidate runs the protocol for a candidate node. Returns the next node type it would like to transition to.
func (state *State) Candidate(sb Switchboard) (newState int) {
	state.incrementTerm()
	state.tryVote(state.id) // Vote for myself (in case anyone asks)
	nVotes := 1

	state.resetElectionTimer()

	// Build request vote query
	request := RequestVoteQ{}
	request.Term = state.currentTerm
	request.CandidateId = state.id
	// TODO: LastLogIndex / LastLogTerm are not used in leader election (what I am focusing on for now...)

	// Send RequestVote to all peers
	state.log("Sending", len(sb.Outs), "RequestVotes at Term ", state.currentTerm)
	outstanding := make(chan *rpc.Call, len(sb.Outs))
	for _, out := range sb.Outs {
		resp := new(RequestVoteA)
		call := out.Go("Node.RequestVote", request, resp, outstanding)
		if call.Error != nil {
			state.log("Error sending RPC: ", call.Error)
		}
	}

	// Handle responses / timeout / losing the election
	state.candidateLost = make(chan struct{})
	for {
		select {
		case <-state.electionTimeout: // timeout, try again
			return Candidate
		case <-state.candidateLost: // we lost the election
			state.log("Election lost.")
			return Follower
		case <-state.becomeFollower: // larger term found
			state.log("Learned of later term number. Forfeiting election.")
			return Follower
		case call := <-outstanding: // response to RequestVote received
			if call.Error != nil {
				state.logger.Printf("rpc error: %v\n", call.Error)
				continue
			}
			resp := call.Reply.(*RequestVoteA)
			if resp.Term > state.currentTerm {
				return Follower
			}

			if resp.VoteGranted {
				nVotes++
				state.log("Received YES vote in Term", resp.Term, "from", strconv.FormatUint(resp.SenderId, 36))
				if nVotes > len(sb.Outs)/2+1 {
					return Leader
				}
			} else {
				state.log("Received NO vote in Term", resp.Term)
			}
			continue
		}
	}
}
