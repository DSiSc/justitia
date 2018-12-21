package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/DSiSc/p2p/tools/common"
	"github.com/DSiSc/p2p/tools/statistics/client"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	var statisticsServer string
	var nodeCount int
	flagSet := flag.NewFlagSet("block-propagator-test", flag.ExitOnError)
	flagSet.StringVar(&statisticsServer, "server", "localhost:8080", "statistics server address, default localhost:8080")
	flagSet.IntVar(&nodeCount, "nodes", 1, "p2p node count")
	flagSet.Usage = func() {
		fmt.Println(`Justitia blockchain block broadcast p2p test tool.
Usage:
	block-propagator-test [-server localhost:8080 -nodes 1]

Examples:
	block-propagator-test -server localhost:8080  -nodes 1`)
		fmt.Println("Flags:")
		flagSet.PrintDefaults()
	}
	flagSet.Parse(os.Args[1:])
	statisticsClient := client.NewStatisticsClient(statisticsServer)

	topo, err := statisticsClient.GetTopos()
	if err != nil {
		fmt.Printf("Failed to get topo info from server, as: %v", err)
		os.Exit(1)
	}
	reachability := client.TopoReachbility(topo)
	if reachability < nodeCount {
		fmt.Printf("The net reachability is %d, less than nodes count： %d\n", reachability, nodeCount)
		os.Exit(1)
	}

	// statistics peer's height
	if heightStatistic(topo) {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

// compare all peer's height
func heightStatistic(topo map[string][]*common.Neighbor) bool {
	currentHeight := -1
	for peer, _ := range topo {
		pH := getPeerHeight(peer)
		if currentHeight >= 0 && pH != uint64(currentHeight) {
			return false
		} else {
			currentHeight = int(pH)
		}
	}
	return true
}

// get peer's height
func getPeerHeight(peer string) uint64 {
	addr := peer[:strings.Index(peer, ":")]
	resp, err := http.Get("http://" + addr + ":" + strconv.Itoa(47768) + "/eth_blockNumber")
	if err != nil {
		fmt.Printf("Failed to get p2p topo info, as: %v\n", err)
		return 0
	}
	var result map[string]string
	parseResp(&result, resp)
	height, _ := strconv.ParseUint(result["result"][2:], 16, 64)
	resp.Body.Close()
	fmt.Printf("%s height: %d\n", addr, height)
	return height
}

func parseResp(v interface{}, resp *http.Response) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body, as: %v\n", err)
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		fmt.Printf("Failed to parse response body, as: %v\n", err)
	}

}
