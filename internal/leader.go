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

// Follower runs the protocol for a leader node. It returns the next node type it would like to transition to.
func (state *State) Leader(sb Switchboard) (newState int) {
	// Initialize leader state
	outstanding := make(chan *rpc.Call, len(sb.Outs))

	for {
		select {
		case <-state.heartbeatTicker:
			state.sendHeartbeats(sb, outstanding)
			continue
		case <-state.becomeFollower:
			state.log("Learned of later term number. Abdicating leadership.")
			return Follower
		case call := <-outstanding:
			if call.Error != nil {
				continue
			}
			resp := call.Reply.(*AppendEntriesA)
			if resp.Term > state.currentTerm {
				return Follower
			}
		}
	}
}

func (state *State) sendHeartbeats(sb Switchboard, outstanding chan *rpc.Call) {
	request := AppendEntriesQ{}
	request.Term = state.currentTerm
	request.Entries = make([]int, 0)
	// TODO: Set other fields as necessary

	var errors = 0
	for _, out := range sb.Outs {
		resp := new(AppendEntriesA)
		call := out.Go("Node.AppendEntries", request, resp, outstanding)
		if call.Error != nil {
			errors++
		}
	}
	if errors > 0 {
		state.log("Couldn't send ", errors, " heartbeats. Assuming follower(s) are crashed.")
	}
}
