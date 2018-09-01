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

	"github.com/kalexmills/elect/internal"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Launches a single elect node",
	Long: `The node listens on the given port and attempts to connect to its peers at the given addresses.
Usage:
    elect launch node [port] [host:port]...

Ex:
    elect launch node 1234 peer01:1234 peer02:1234 peer03:1234
`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		port, e := strconv.ParseUint(args[0], 10, 32)
		if e != nil {
			fmt.Println("First argument must be parseable as an integer.")
			os.Exit(1)
		}

		elect.Launch(port, args[1:])
	},
}

func init() {
	launchCmd.AddCommand(nodeCmd)
}
