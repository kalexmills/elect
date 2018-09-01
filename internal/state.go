package elect

import (
	"log"
	"math/rand"
	"net/rpc"
	"time"
	"fmt"
)

const (
	Leader = iota + 1
	Follower
	Candidate
)

// State stores the state of a single node in the Raft protocol.
type State struct {
	PersistentState
	VolatileState
	LeaderState
}

// TODO: Encapsulate persistence
type PersistentState struct {
	currentTerm uint64
	votedFor    uint64
	log         []uint
}

type VolatileState struct {
	commitIndex uint64
	lastApplied uint64

	id            uint64
	state         int

	logger *log.Logger

	electionTimer *time.Timer

	// electionTimeout occurs after not receiving a message for a period
	electionTimeout chan struct{}
	// candateLost occurs if the candidate loses
	candidateLost chan struct{}
	// becomeFollower occurs when a Term higher than the current Term is found
	becomeFollower chan struct{}
}

type LeaderState struct {
	nextIndex  []uint
	matchIndex []uint

	heartbeatTicker <-chan time.Time
}

// Launch launches this node, listening on a TCP connection on the given port and attempting to connect to
// the given peers, which must be specified as "host:port" strings. The function blocks as long as the
// raft node is running.
func Launch(port uint64, peers []string) {
	// Register a new RPC handler
	node := new(Node)
	node.localstate = new(State)
	rpc.Register(node)

	node.localstate.Raft(port, peers)
}

func (s *PersistentState) incrementTerm() {
	s.currentTerm++
	s.votedFor = Noone
}

// tryVote attempts to vote for the node with id. The result is the node which has been voted for in this Term.
func (s *PersistentState) tryVote(id uint64) uint64 {
	if s.votedFor == Noone {
		s.votedFor = id
	}
	return s.votedFor
}

// onReceiveRpc encapsulates the common activities each node must perform when they receive an RPC
func (state *State) onReceiveRpc(term uint64) {
	state.resetElectionTimer()
	state.maybeSignalTermExceeded(term)
}

// resetElectionTimer sets the election timer to a random timeout between MinElectionTimeout and MaxElectionTimeout
func (state *State) resetElectionTimer() {
	delay := time.Millisecond * time.Duration(rand.Intn(MaxElectionTimeout-MinElectionTimeout)+MinElectionTimeout)

	//state.log("Setting election timeout to ", delay)
	if state.electionTimer == nil || state.electionTimer.Stop() {
		state.electionTimer = time.AfterFunc(delay, func() {
			state.electionTimeout <- struct{}{}
		})
	}
}

// maybeSignalTermExceeded checks to see if the Term has been exceeded and, if so, signals that this node should become
// a follower.
func (state *State) maybeSignalTermExceeded(newTerm uint64) {
	if state.currentTerm < newTerm {
		state.currentTerm = newTerm
		if state.state != Follower {
			state.becomeFollower <- struct{}{}
		}
	}
}

// signalCandidateLost signals that the candidate has lost
func (state *State) signalCandidateLost() {
	if state.state == Candidate {
		state.candidateLost <- struct{}{}
	}
}

func (state *State) log(msg ...interface{}) {
	var s string
	switch state.state {
	case Leader: s =    "LEADER"
		break
	case Follower: s =  "FOLLOW"
		break
	case Candidate: s = "CANDID"
		break
	}
	state.logger.Printf("%s %4d %s", s, state.currentTerm, fmt.Sprintln(msg...))
}
