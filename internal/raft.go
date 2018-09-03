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
	crypto "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"time"
)

const (
	MinElectionTimeout = 150
	MaxElectionTimeout = 300
	HeartbeatTimeout   = MinElectionTimeout - 50 // adjust based on your own RTT and clock-sync assumptions
	Noone              = 0
)

// raft sets up the initial volatile role and
func (state *state) raft(port uint64, peers []string) {
	defer func() {
		if err := recover(); err != nil {
			state.logger.Printf("%d panicking! err: %v", state.id, err)
			panic(err)
		}
	}()

	// Setup memory
	state.votedFor = Noone

	// Seed RNG because timeouts need to be "random enough".
	seed, _ := crypto.Int(crypto.Reader, big.NewInt(int64(^uint64(0)>>1)))
	rand.Seed(seed.Int64())

	state.id = uint64(os.Getpid())

	state.logger = log.New(os.Stdout, fmt.Sprintf("%-5d ", state.id), log.Lmicroseconds)

	n := len(peers)
	state.electionTimeout = make(chan struct{})
	state.becomeFollower = make(chan struct{}, n)
	state.candidateLost = make(chan struct{}, n)

	state.heartbeatTicker = time.Tick(HeartbeatTimeout * time.Millisecond)

	state.role = Follower

	var sb Switchboard
	sb.Initialize(port, peers)

	for {
		// Each role only takes on one role at a time. When one of the Run* functions
		// needs to change roles, it just returns the role it is transitioning to.
		switch state.role {
		case Follower:
			state.log("Transitioning to Follower")
			state.role = state.Follower()
			continue
		case Candidate:
			state.log("Transitioning to candidate")
			state.role = state.candidate(sb)
			continue
		case Leader:
			state.log("Transitioning to leader")
			state.role = state.leader(sb)
		}
	}
}
