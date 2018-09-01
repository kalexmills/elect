// Copyright © 2018 K. Alex Mills <k.alex.mills@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package elect

import (
	"net/rpc"
)

// candidate runs the protocol for a candidate node. Returns the next node type it would like to transition to.
func (state *State) candidate(sb Switchboard) (newState int) {
	state.incrementTerm()

	// Candidates vote for themselves (in case anyone asks)
	state.tryVote(state.id)
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
				state.log("Received YES vote in Term", resp.Term, "from", resp.SenderId)
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
