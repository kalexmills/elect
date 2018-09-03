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
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"time"
)

const (
	Leader = iota + 1
	Follower
	Candidate
)

// state stores the state of a single node in the raft protocol.
type state struct {
	persistentState
	volatileState
	leaderState
}

// TODO: Encapsulate persistence
type persistentState struct {
	currentTerm uint64
	votedFor    uint64
	log         []string
}

// volatileState stores the state which is not saved to disk.
type volatileState struct {
	commitIndex uint64
	lastApplied uint64

	id    uint64
	state int

	logger *log.Logger

	electionTimer *time.Timer

	// electionTimeout occurs after not receiving a message for a period
	electionTimeout chan struct{}
	// candateLost occurs if the candidate loses
	candidateLost chan struct{}
	// becomeFollower occurs when a Term higher than the current Term is found
	becomeFollower chan struct{}
}

// leaderState stores state which is only allocated when a node becomes a leader.
type leaderState struct {
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
	node.localstate = new(state)
	rpc.Register(node)

	node.localstate.raft(port, peers)
}

func (s *persistentState) incrementTerm() {
	s.currentTerm++
	s.votedFor = Noone
}

// tryVote attempts to vote for the node with id. The result is the node which has been voted for in this Term.
func (s *persistentState) tryVote(id uint64) uint64 {
	if s.votedFor == Noone {
		s.votedFor = id
	}
	return s.votedFor
}

// onReceiveRpc encapsulates the common activities each node must perform when they receive an RPC
func (state *state) onReceiveRpc(term uint64) {
	state.resetElectionTimer()
	state.maybeSignalTermExceeded(term)
}

// resetElectionTimer sets the election timer to a random timeout between MinElectionTimeout and MaxElectionTimeout
func (state *volatileState) resetElectionTimer() {
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
func (state *state) maybeSignalTermExceeded(newTerm uint64) {
	if state.currentTerm < newTerm {
		state.currentTerm = newTerm
		if state.state != Follower {
			state.becomeFollower <- struct{}{}
		}
	}
}

// signalCandidateLost signals that the candidate has lost
func (state *volatileState) signalCandidateLost() {
	if state.state == Candidate {
		state.candidateLost <- struct{}{}
	}
}

func (state *state) log(msg ...interface{}) {
	var s string
	switch state.state {
	case Leader:
		s = "LEADER"
		break
	case Follower:
		s = "FOLLOW"
		break
	case Candidate:
		s = "CANDID"
		break
	}
	state.logger.Printf("%s %4d %s", s, state.currentTerm, fmt.Sprintln(msg...))
}
