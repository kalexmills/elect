package elect

import (
	"time"
	"net/rpc"
	"log"
	"math/rand"
)

const (
	Leader = iota + 1
	Follower
	Candidate
)

// Node serves as a public interface for RPC calls.
type Node struct {
	localstate *State
}

// State stores the state of a single node in the Raft protocol.
type State struct {
	PersistentState
	VolatileState
	LeaderState
}

// TODO: Encapsulate persistence
type PersistentState struct {
	currentTerm uint64
	votedFor uint64
	log []uint
}

type VolatileState struct {
	commitIndex uint64
	lastApplied uint64

	id uint64
	state int
	timer *time.Timer

	logger *log.Logger

	// candateLost is closed if the candidate loses
	candidateLost chan struct{}
	// becomeFollower is closed when a Term higher than the current Term is found
	becomeFollower chan struct{}
}

type LeaderState struct {
	nextIndex []uint
	matchIndex []uint
}

// Launch launches this node, listening on a TCP connection on the given port and attempting to connect to
// the given peers, which must be specified as "host:port" strings. The function blocks as long as the
// raft node is running.
func Launch(port uint64, peers []string) {
	// Register a new RPC handler
	node := new(Node)
	node.localstate = new(State)
	rpc.Register(node)

	var net Switchboard
	net.Initialize(port, peers)

	node.localstate.RunRaft(net)
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