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
	"net"
	"net/rpc"
	"sync"
)

// Switchboard is a structure used to persist the local state of network connections among peers.
type Switchboard struct {
	In   net.Listener
	Outs []*rpc.Client
}

func (sb *Switchboard) Initialize(port uint64, peers []string) {
	n := len(peers)
	var wg sync.WaitGroup
	wg.Add(n + 1)

	// Initialize receiver
	go func(port uint64) {
		defer wg.Done()
		var e error

		sb.In, e = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if e != nil {
			log.Fatalf("listen error: %v", e)
		}
		go func() {
			for {
				conn, _ := sb.In.Accept()
				go rpc.ServeConn(conn)
			}
		}()
	}(port)

	// Initialize senders in parallel
	sb.Outs = make([]*rpc.Client, n)
	for i := 0; i < len(peers); i++ {
		go func(peer string, i int, outs []*rpc.Client) {
			defer wg.Done()

			var e error
			outs[i], e = rpc.Dial("tcp", peer)
			if e != nil {
				log.Fatal("dial error:", e)
			}
		}(peers[i], i, sb.Outs)
	}
	wg.Wait()
}
