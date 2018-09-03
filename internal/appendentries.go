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
	// Term is the receiver's currentTerm
	Term    uint64
	lastLogIndex uint64
	Success bool
}

// AppendEntries is the RPC Handler that handles AppendEntries calls
func (node *Node) AppendEntries(in AppendEntriesQ, out *AppendEntriesA) error {
	state := node.localstate
	state.resetElectionTimer()

	if in.Term < state.currentTerm {
		// Reject all AppendEntries messages whose terms come before our own.
		out.Success = false
		out.Term = state.currentTerm
		return nil
	}
	if state.currentTerm < in.Term {
		// Update our term and revert to follower status.
		if state.role == Candidate {
			state.signalCandidateLost()
		}
		state.votedFor = Noone
		state.currentTerm = in.Term
		if state.role != Follower {
			state.becomeFollower <- struct{}{}
		}
	}

	// Quit early on a heartbeat.
	if len(in.Entries) == 0 {
		return nil
	}

	state.log("AppendEntries message received during Term ", in.Term)

	out.Term = state.currentTerm

	// For leader-election, all that is needed is that this serves as a heart-beat.
	return nil
}
