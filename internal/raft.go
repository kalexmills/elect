package elect

import (
	crypto "crypto/rand"
	"math/big"
	"math/rand"
	"strconv"
	"log"
	"os"
	"fmt"
	"time"
)

const (
	MinElectionTimeout = 150
	MaxElectionTimeout = 300
	HeartbeatTimeout   = MinElectionTimeout - 50 // adjust based on your own RTT and clock-sync assumptions
	Noone              = 0
)

func (state *State) Raft(port uint64, peers []string) {
	defer func() {
		if err := recover(); err != nil {
			state.logger.Printf("panicking! err: %v", strconv.FormatUint(state.id, 36), err)
			panic(err)
		}
	}()

	// Setup memory
	state.votedFor = Noone

	state.id = Noone // 0 is a special ID used to indicate that this state hasn't (yet) cast a vote
	seed, _ := crypto.Int(crypto.Reader, big.NewInt(int64(^uint64(0)>>1)))
	rand.Seed(seed.Int64())
	for state.id == Noone {
		state.id = rand.Uint64() // New state; new ID. This means a crashed candidate definitely loses its election.
	}
	state.logger = log.New(os.Stdout, fmt.Sprintf("%-13s ", strconv.FormatUint(state.id, 36)), log.Lmicroseconds)

	n := len(peers)
	state.electionTimeout = make(chan struct{})
	state.becomeFollower = make(chan struct{}, n)
	state.candidateLost = make(chan struct{}, n)

	state.heartbeatTicker = time.Tick(HeartbeatTimeout * time.Millisecond)

	state.state = Follower

	var sb Switchboard
	sb.Initialize(port, peers)

	for {
		// Each state only takes on one role at a time. When one of the Run* functions
		// needs to change roles, it just returns the role it is transitioning to.
		switch state.state {
		case Follower:
			state.log("Transitioning to Follower")
			state.state = state.Follower()
			continue
		case Candidate:
			state.log("Transitioning to Candidate")
			state.state = state.Candidate(sb)
			continue
		case Leader:
			state.log("Transitioning to Leader")
			state.state = state.Leader(sb)
		}
	}
}