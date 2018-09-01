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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"os"
	"strconv"
	"sync"
	"os/exec"
	"os/signal"
)

// localCmd represents the local command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Launches a cluster of local elect nodes.",
	Long: `Launches multiple elect nodes running as separate goroutines. Each listens to one of the given ports and
connects to one another via localhost.

Usage:
    elect launch cluster [port]...`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ports := make([]uint64, len(args))
		for i, arg := range args {
			var e error
			ports[i], e = strconv.ParseUint(arg, 10, 32)
			if e != nil {
				fmt.Println("Each argument must be parseable as an integer.")
				os.Exit(1)
			}
		}

		// Runs each node in a separate goroutine.
		var wg sync.WaitGroup
		wg.Add(len(ports))

		stopRestarting := make(chan struct{})
		running := make(chan *exec.Cmd, len(ports))
		for i := 0; i < len(ports); i++ {

			k := 0
			peers := make([]string, len(ports) - 1)
			for j := 0; j < len(ports); j++ {
				if ports[i] != ports[j] {
					peers[k] = fmt.Sprintf("localhost:%d", ports[j])
					k++
				}
 			}

 			// Starts a cluster node as a local process, and restarts it when it crashes.
			go func(port uint64, peers []string) {

				for {
					cmd := exec.Command(".\\elect", append([]string{"launch","node", strconv.FormatUint(port, 10)}, peers...)...)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr

					running <- cmd
					err := cmd.Run()
					if err != nil {
						fmt.Println("Could not start subprocess:", err)
					}

					// Keep the process if it crashes, but if this process is killed, stop rebooting them.
					terminated := make(chan struct{})
					go func() {
						cmd.Wait()
						terminated <- struct{}{}
					}()

					select {
						case <-terminated:
							continue
						case <- stopRestarting:
							break
					}
				}

				wg.Done()
			}(ports[i], peers)
		}

		// On interrupt, kill all subprocesses before exit.

		interrupt := make(chan os.Signal, 2)
		signal.Notify(interrupt, os.Interrupt, os.Kill)
		go func() {
			<-interrupt
			stopRestarting <- struct{}{}
			for cmd := range running {
				if err := cmd.Process.Kill(); err != nil {
					fmt.Println("Failed to kill subprocess ", cmd.Process.Pid, " error ", err)
				}
			}
			os.Exit(1)
		}()
		fmt.Println("Cluster created.")
		wg.Wait() // This waitgroup never halts unless someone killed all our nodes {o.o}
	},
}

func init() {
	launchCmd.AddCommand(clusterCmd)
}
