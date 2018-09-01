package elect

import (
	"net"
	"net/rpc"
	"fmt"
	"log"
	"sync"
	"os"
)

// Switchboard is a structure used to persist the local state of network connections among peers.
type Switchboard struct {
	In net.Listener
	Outs []*rpc.Client
}

func (sb *Switchboard) Initialize(port uint64, peers []string) {
	logger := log.New(os.Stdout, fmt.Sprintf("port %d switchboard  ", port), log.Lmicroseconds)
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
		logger.Println("Listening on port ", port)
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
			logger.Println("Dialed", peer, "successfully")
		}(peers[i], i, sb.Outs)
	}
	wg.Wait()
}