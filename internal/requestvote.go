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
	state.resetElectionTimer()
	if state.currentTerm < in.Term {
		if state.role != Follower {
			state.votedFor = Noone // Reset votedFor so the response to this RPC is a vote for the candidate
		}
		state.currentTerm = in.Term
		if state.role != Follower {
			state.becomeFollower <- struct{}{}
		}
	}

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
