package elect

import (
	"time"
	"net/rpc"
	"log"
	"fmt"
	"os"
	"strconv"
	"math/rand"
	"math/big"
	crypto "crypto/rand"
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

	seed, _ := crypto.Int(crypto.Reader, big.NewInt(int64(^uint64(0) >> 1)))
	rand.Seed(seed.Int64())
	for state.id == Noone {
		state.id = rand.Uint64() // New state; new ID. This means a crashed candidate definitely loses its election.
	}

	n := len(sb.Outs)
	state.becomeFollower = make(chan struct{}, n)
	state.candidateLost = make(chan struct{}, n)

	state.state = Follower
	state.timer = time.NewTimer(1 * time.Second) // RunFollower will reset this

	state.logger = log.New(os.Stdout, fmt.Sprintf("%-13s ", strconv.FormatUint(state.id, 36)), log.Lmicroseconds)

	for {
		// Each state only takes on one role at a time. When one of the Run* functions
		// needs to change roles, it just returns the role it is transitioning to.
		switch state.state {
			case Follower:
				state.logger.Println("Transitioning to Follower")
				state.state = state.RunFollower(sb)
				continue
			case Candidate:
				state.logger.Println("Transitioning to Candidate")
				state.state = state.RunCandidate(sb)
				continue
			case Leader:
				state.logger.Println("Transitioning to Leader")
				state.state = state.RunLeader(sb)
		}
	}
}

// RunFollower runs the protocol for a follower node. Returns the next node type it would like to transition to.
func (state *State) RunFollower (sb Switchboard) (newState int) {
	state.ResetTimer()
	state.votedFor = Noone // Brand new followers just vote for the first leader they hear from.

	// Do nothing on timeout except become a candidate
	for {
		select {
			case <-state.timer.C:
				state.logger.Println("Election timeout occurred")
				return Candidate
			case <-state.becomeFollower: // Clear the becomeFollower channel, ignoring any extra messages
				continue
		}
	}
}

// RunCandidate runs the protocol for a candidate node. Returns the next node type it would like to transition to.
func (state *State) RunCandidate(sb Switchboard) (newState int){
	state.incrementTerm()
	state.tryVote(state.id) // Vote for myself (in case anyone asks)
	nVotes := 1

	// Build request vote query
	request := RequestVoteQ{}
	request.Term = state.currentTerm
	request.CandidateId = state.id
	// TODO: LastLogIndex / LastLogTerm are not used in leader election (what I am focusing on for now...)

	// Send RequestVote to all peers
	state.logger.Println("Sending", len(sb.Outs) ,"RequestVotes at Term ", state.currentTerm)
	outstanding := make(chan *rpc.Call, len(sb.Outs))
	for _, out := range sb.Outs {
		resp := new(RequestVoteA)
		call := out.Go("Node.RequestVote", request, resp, outstanding)
		if call.Error != nil {
			state.logger.Println("Error sending RPC: ", call.Error)
		}
	}

	// Handle responses / timeout / losing the election
	state.candidateLost = make(chan struct{})
	for {
		select {
			case <-state.timer.C: // timeout
				// Start a new election by becoming a candidate again.
				return Candidate
			case <-state.candidateLost: // we lost the election
				state.logger.Println("Election lost.")
				return Follower
			case <-state.becomeFollower: // larger Term found -- we REALLY lost the election
				state.logger.Println("Learned of later term number. Forfeiting election.")
				return Follower
			case call := <-outstanding: // response
				if call.Error != nil {
					state.logger.Printf("error: %v\n", call.Error)
				}
				resp := call.Reply.(*RequestVoteA)
				if resp.Term > state.currentTerm { return Follower }

				if resp.VoteGranted {
					nVotes++
					state.logger.Println("Received YES vote in Term", resp.Term, "from", resp.SenderId)
					if nVotes > len(sb.Outs) / 2 + 1 {
						return Leader
					}
				} else {
					state.logger.Println("Received NO vote in Term", resp.Term)
				}
				continue
		}
	}
}

// RunFollower runs the protocol for a leader node. It returns the next node type it would like to transition to.
func (state *State) RunLeader(sb Switchboard) (newState int) {
	state.logger.Println("Leadership achieved")
	// Leader holds onto their power forever. It's good to be king!
	for {
		select {
		case <-state.becomeFollower:
			state.logger.Println("Learned of later term number. Abdicating leadership.")
			return Follower
		}
	}
}

func (state *State) ResetTimer() {
	delay := time.Duration(rand.Intn(MaxElectionTimeout - MinElectionTimeout) + MinElectionTimeout) * time.Millisecond
	state.timer.Reset(delay)
	state.logger.Println("Setting election timeout to ", delay)
}

// OnReceiveRPC encapsulates the common activities each node must perform when they receive an RPC
func (state *State) OnReceiveRPC(term uint64) {
	state.ResetTimer()
	state.MaybeSignalTermExceeded(term)
}

// MaybeSignalTermExceeded checks to see if the Term has been exceeded and, if so, signals that this node should become
// a follower.
func (state *State) MaybeSignalTermExceeded(newTerm uint64) {
	if state.currentTerm < newTerm {
		state.currentTerm = newTerm
		state.becomeFollower <- struct{}{}
	}
}

func (state *State) SignalCandidateLost() {
	state.candidateLost <- struct{}{}
}