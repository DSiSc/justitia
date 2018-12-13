package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/p2p/tools/common"
	"github.com/DSiSc/p2p/tools/statistics/client"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// RPCError rpc response error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// SendTx response
type SendTxResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  string      `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// TxResponse get transaction response
type TxResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Error   *RPCError   `json:"error,omitempty"`
}

func main() {
	var statisticsServer string
	var nodeCount int
	flagSet := flag.NewFlagSet("block-propagator-test", flag.ExitOnError)
	flagSet.StringVar(&statisticsServer, "server", "localhost:8080", "statistics server address, default localhost:8080")
	flagSet.IntVar(&nodeCount, "nodes", 20, "p2p node count")
	flagSet.Usage = func() {
		fmt.Println(`Justitia blockchain tx propagator test client.
Usage:
	block-propagator-test [-server localhost:8080 -nodes 1 ]

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

	// send tx and check the result
	txs := sendTxs(topo)
	checkTxs(txs, topo)

	// report the longest time that transaction broadcast
	checkLongestBroadcastTime(txs, statisticsClient)
}

// send transaction to peer
func sendTxs(topo map[string][]*common.Neighbor) []string {
	txs := make([]string, 0)
	for peer, _ := range topo {
		reqData := fmt.Sprintf(
			`{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from": "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b","to": "0xa94f5374fce5edbc8e2a8697c15331677e6ebf01","gas": "0x6400","gasPrice": "0x1234","value": "0x%x"}],"id":1}`, rand.Intn(100000000))
		addr := peer[:strings.Index(peer, ":")]
		resp, err := http.Post("http://"+addr+":47768", "application/json", strings.NewReader(reqData))
		if err != nil {
			fmt.Printf("New request error, please check. error info: %v\n", err)
			os.Exit(1)
		}

		result := new(SendTxResponse)
		parseResp(result, resp)
		resp.Body.Close()
		txs = append(txs, result.Result)
	}
	return txs
}

// check transaction
func checkTxs(txs []string, topo map[string][]*common.Neighbor) {
	for _, tx := range txs {
		for peer, _ := range topo {
			reqData := fmt.Sprintf(`{"jsonrpc": "2.0","method": "eth_getTransactionByHash","params": ["%s"],"id": 1}`, tx)
			addr := peer[:strings.Index(peer, ":")]
			resp, err := http.Post("http://"+addr+":47768", "application/json", strings.NewReader(reqData))
			if err != nil {
				log.Error("New request error, please check.")
				os.Exit(3)
			}

			result := new(TxResponse)
			parseResp(&result, resp)
			resp.Body.Close()
			if result.Error != nil {
				fmt.Printf("node %s have not tx %s", peer, tx)
				os.Exit(4)
			}
		}
	}
}

// check the longest broadcast time
func checkLongestBroadcastTime(txs []string, statisticsClient *client.StatisticsClient) {
	for _, tx := range txs {
		msgReport, err := statisticsClient.GetReportMessage(tx)
		if err != nil {
			fmt.Printf("failed to get tx %s's report info, as: %v", tx, err)
			os.Exit(1)
		}
		var timeStart, timeEnd *time.Time
		for _, report := range msgReport {
			if timeStart == nil || report.Time.Before(*timeStart) {
				timeStart = &report.Time
			}
			if timeEnd == nil || report.Time.After(*timeEnd) {
				timeEnd = &report.Time
			}
		}
		fmt.Printf("Tx %s broadcast time is %v\n", tx, timeEnd.Sub(*timeStart))
	}
}

// unmarshal response to specified type
func parseResp(v interface{}, resp *http.Response) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body, as: %v\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		fmt.Printf("Failed to parse response body, as: %v\n", err)
		os.Exit(1)
	}

}
