package elect

import (
	"golang.org/x/exp/rand"
	"time"
	"net/rpc"
)

const (
	MinElectionTimeout = 150
	MaxElectionTimeout = 300
	Noone             = 0
)

func (state *State) RunRaft(sb Switchboard) {
	// Setup memory
	state.votedFor = Noone

	state.id = Noone // 0 is a special ID used to indicate that this state hasn't (yet) cast a vote
	for state.id == Noone {
		state.id = rand.Uint64() // New state; new ID. This means a crashed candidate definitely loses its election.
	}

	state.state = Follower
	state.timer = time.NewTimer(1 * time.Second) // RunFollower will reset this

	for {
		// Each state only takes on one role at a time. When one of the Run* functions
		// needs to change roles, it just returns the role it is transitioning to.
		switch state.state {
			case Follower:
				state.state = state.RunFollower(sb)
				continue
			case Candidate:
				state.state = state.RunCandidate(sb)
				continue
			case Leader:
				state.state = state.RunLeader(sb)
		}
	}
}

// RunFollower runs the protocol for a follower node. Returns the next node type it would like to transition to.
func (state *State) RunFollower (sb Switchboard) (newState int) {

	state.ResetTimer()

	// Do nothing on timeout except become a candidate
	select {
		case <-state.timer.C:
			return Candidate
	}
}

// RunCandidate runs the protocol for a candidate node. Returns the next node type it would like to transition to.
func (state *State) RunCandidate(sb Switchboard) (newState int){
	state.incrementTerm()
	state.tryVote(state.id) // Vote for myself (in case anyone asks)
	nVotes := 1

	state.ResetTimer()

	// Build request vote query
	request := RequestVoteQ{}
	request.term = state.currentTerm
	request.candidateId = state.id
	// TODO: lastLogIndex / lastLogTerm are not used in leader election (what I am focusing on for now...)

	// Send RequestVote to all peers
	outstanding := make(chan *rpc.Call, len(sb.Outs))
	for _, out := range sb.Outs {
		resp := new(RequestVoteA)
		out.Go("State.RequestVote", request, resp, outstanding)
	}

	// Handle responses / timeout / losing the election
	state.candidateLost = make(chan struct{})
	for {
		select {
			case call := <- outstanding: // response
				resp := call.Reply.(*RequestVoteA)
				if resp.term > state.currentTerm { return Follower }

				if resp.voteGranted {
					nVotes++
					if nVotes > len(sb.Outs) / 2 + 1 {
						return Leader
					}
				}
			case <-state.timer.C: // timeout
				// Start a new election by becoming a candidate again.
				return Candidate
			case <-state.candidateLost: // we lost the election
			case <-state.becomeFollower: // larger term found -- we REALLY lost the election
				return Follower
		}
	}
}

// RunFollower runs the protocol for a leader node. It returns the next node type it would like to transition to.
func (state *State) RunLeader(sb Switchboard) (newState int) {
	return Leader // Always be the king
	// TODO: Send heartbeats (at least)
}

func (state *State) ResetTimer() {
	state.timer.Reset(time.Duration(rand.Intn(MaxElectionTimeout-MinElectionTimeout) +MinElectionTimeout) * time.Millisecond)
}

// OnReceiveRPC encapsulates the common activities each node must perform when they receive an RPC
func (state *State) OnReceiveRPC(term uint64) {
	state.ResetTimer()
	state.MaybeSignalCandidateLost()
	state.MaybeSignalTermExceeded(term)
}

// MaybeSignalTermExceeded checks to see if the term has been exceeded and, if so, signals that this node should become
// a follower.
func (state *State) MaybeSignalTermExceeded(newTerm uint64) {
	if state.currentTerm < newTerm {
		close(state.becomeFollower)
		state.becomeFollower = make(chan struct{})
	}
}

func (state *State) MaybeSignalCandidateLost() {
	if state.candidateLost != nil {
		close(state.candidateLost)
		state.candidateLost = nil
	}
}
