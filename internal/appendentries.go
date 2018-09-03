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
	Entries      []LogEntry
	LeaderCommit uint64
}

// AppendEntriesA is the response to AppendEntries messages.
type AppendEntriesA struct {
	// Term is the receiver's currentTerm
	Term    uint64
	// Success is true iff the RPC resulted in all entries being successfully appended to remote log
	Success bool
}

// AppendEntries is the RPC Handler that handles AppendEntries calls
func (node *Node) AppendEntries(in AppendEntriesQ, out *AppendEntriesA) error {
	state := node.localstate
	state.resetElectionTimer()

	if in.Term < state.currentTerm {
		// Reject all AppendEntries messages whose terms come before our own.
		return sendFailure(state.currentTerm, out)
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
	if state.entries[in.PrevLogIndex].term != in.PrevLogTerm {
		// Reject messages whose previous index and term do not match our own.
		return sendFailure(state.currentTerm, out)
	}

	// Quit early on a heartbeat.
	if len(in.Entries) == 0 {
		return sendSuccess(state.currentTerm, out)
	}

	// Check for conflicts with previous entries and toss out conflicting entries if needed.
	start := in.PrevLogIndex
	i := uint64(0)
	for j := range in.Entries {
		if state.entries[start + i].term != in.Entries[j].term {
			state.entries = state.entries[:i]
		}
		i++
	}

	// Append new entries not already in log.
	state.entries = append(state.entries, in.Entries[i:]...)

	// Update commit index as needed.
	if in.LeaderCommit > state.commitIndex {
		state.commitIndex = min(in.LeaderCommit, uint64(len(state.entries)))
	}

	state.log("AppendEntries message received during Term ", in.Term)

	return sendSuccess(state.currentTerm, out)
}

func min(a uint64, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

func sendFailure(term uint64, out *AppendEntriesA) error {
	out.Success = false
	out.Term = term
	return nil
}

func sendSuccess(term uint64, out *AppendEntriesA) error {
	out.Success = true
	out.Term = term
	return nil
}