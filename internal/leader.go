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
