package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/DSiSc/craft/log"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}
type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  string      `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

var (
	client    = &http.Client{}
	endpoints []string
	port string
)

// main process goes here.
func main() {
	var durationInt, txsRate, blockStart, blockEnd uint
	var verbose, random, showHelp, statistic bool
	var timerWG sync.WaitGroup
	var filePath string

	//////////////////////////////////
	// flagSet handles command flags.
	//////////////////////////////////
	flagSet := flag.NewFlagSet("jt-bench", flag.ExitOnError)
	flagSet.UintVar(&durationInt, "t", 60, "Exit after the specified amount of time in seconds(0 means forever).")
	flagSet.UintVar(&txsRate, "r", 200, "Txs per second to send in a connection.")
	flagSet.StringVar(&port, "p", "47768", "Port of api gateway (default to 47768).")
	flagSet.BoolVar(&verbose, "v", false, "Verbose output.")
	flagSet.BoolVar(&showHelp, "h", false, "Display help.")
	flagSet.BoolVar(&random, "R", false, "Random number of tx.")
	flagSet.BoolVar(&statistic, "s", false, "The switch of statistic and bench.")
	flagSet.UintVar(&blockStart, "b", 1, "The number of starting statistics.")
	flagSet.UintVar(&blockEnd, "e", 1, "The number of ending statistics.")
	flagSet.StringVar(&filePath, "f", "/var/log/justitia/justitia.log", "The absolute path of statistic file.")
	flagSet.Usage = func() {
		fmt.Println(`Justitia blockchain benchmarking tool.

Usage:
    Send TXs:  go run bench.go [-t 60] [-r 200] [-p 47768] [-R] [-v] [host1,host2,host3,host4]
    Statistic: go run bench.go -s -b 100 -e 200 [-p 47768] [-v] [host]

Examples:
    There are two ways to use this tools.

    First is to bench, simulate send 200 txs/s in 60 second
    It will output starting, ending block number and tps, bps at the same time.

        go run bench.go -t 60 -r 200 -v 192.168.1.100,192.168.1.101,192.168.1.102,192.168.1.103

        Or keep sending random number of tx forever

        go run bench.go -t 0 -R

    Second is statistic output of the bench result, it will output log info
    when making block, and tps, bps at the same time.

        go run bench.go -s -b 120 -e 160

   `)

		fmt.Println("Flags:")
		flagSet.PrintDefaults()
	}
	flagSet.Parse(os.Args[1:])

	if showHelp {
		flagSet.Usage()
		return
	}

	if flagSet.NArg() == 0 {
		endpoints = []string{"127.0.0.1"}
	} else {
		endpoints = strings.Split(flagSet.Arg(0), ",")
	}

	//////////////////////////////////
	// log configuration.
	//////////////////////////////////
	log.RemoveAppender("stdout")
	log.AddAppender("console", os.Stdout, log.DebugLevel, log.TextFmt, false, false)
	if verbose {
		log.SetGlobalLogLevel(log.DebugLevel)
	} else {
		log.SetGlobalLogLevel(log.InfoLevel)
	}

	//////////////////////////////////
	// statistic and report.
	//////////////////////////////////
	if statistic {
		statisticBetweenBlock(int64(blockStart), int64(blockEnd))
		return
	}

	//////////////////////////////////
	// preparation
	//////////////////////////////////
	timerWG.Add(2)
	var ticker, totalTimer *time.Ticker
	var progress []byte
	stopped := false
	showProgress := true

	if durationInt == 0 {
		durationInt = math.MaxInt32
		showProgress = false
	} else if durationInt < 10 {
		log.Error("Duration too short, recommend a bigger integer.")
		return
	} else if durationInt > 3600 {
		showProgress = false
	}

	if showProgress {
		progress = make([]byte, durationInt)
		for i := range progress {
			progress[i] = ' '
		}
	}

	//////////////////////////////////
	// send tx at a given rate
	//////////////////////////////////
	go func() {
		log.Info(fmt.Sprintf("Sending TX at rate of %d tx/sec for %d sec ...", txsRate, durationInt))
		ticker = time.NewTicker(1 * time.Second)
		sent := 0
		for i := 0; i < int(durationInt); i++ {
			go func() {
				if random {
					count := rand.Intn(int(txsRate) * 2)
					log.Debug(fmt.Sprintf("Sending %d TXs ...", count))
					sendTXs(count)
				} else {
					temp := i
					if showProgress {
						progress[temp] = '-'
						fmt.Fprintf(os.Stdout, "Sending... |%s| %2d%% started, %2d%% finished.\r", string(progress[:]), (i+1)*100/int(durationInt), (sent+1)*100/int(durationInt))
					}

					sendTXs(int(txsRate))

					sent = temp
					if showProgress && !stopped {
						progress[temp] = '='
						fmt.Fprintf(os.Stdout, "Sending... |%s| %2d%% started, %2d%% finished.\r", string(progress[:]), (i+1)*100/int(durationInt), (sent+1)*100/int(durationInt))
					}
				}
			}()
			<-ticker.C
		}
		fmt.Println()
		timerWG.Done()
	}()

	////////////////////////////////////////////////////////////////
	// accumulate tx number of blocks generated during bench period.
	////////////////////////////////////////////////////////////////
	var beginningHeight, endingHeight int64
	go func() {
		totalTimer = time.NewTicker(time.Second * time.Duration(durationInt))

		beginningHeight = latestBlockNumber() + 1
		<-totalTimer.C
		endingHeight = latestBlockNumber()

		stopped = true
		timerWG.Done()
	}()

	timerWG.Wait()

	statisticBetweenBlock(beginningHeight, endingHeight)
}

// latestBlockNumber fetches current block number.
func latestBlockNumber() int64 {
	reqData := `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":8}`
	recv := doPost(reqData)
	return hexstr2dec(recv.Result)
}

// timestampOfBlockByNumber returns the time when specified block was created.
func timestampOfBlockByNumber(blockNum int64) int64 {
	reqData := fmt.Sprintf(`{"jsonrpc": "2.0", "method": "eth_getBlockByNumber", "id": 1, "params": ["0x%x", true]}`, blockNum)
	resp, err := http.Post(randomEndpoint(), "application/json", strings.NewReader(reqData))
	if err != nil {
		panic(nil)
	}
	defer resp.Body.Close()
	blob, _ := ioutil.ReadAll(resp.Body)
	bodystr := string(blob)
	index := strings.Index(bodystr, "timestamp")
	return hexstr2dec(bodystr[index+13 : index+23])
}

func randomEndpoint() string {
	return "http://" + endpoints[rand.Intn(len(endpoints))] + ":" + port
}

// txNumOfBlock gets tx number of given block number.
func txNumOfBlock(blockNum int64) int64 {
	reqData := fmt.Sprintf(`{"jsonrpc": "2.0", "method": "eth_getBlockTransactionCountByNumber", "id": 1, "params": ["0x%x"]}`, blockNum)
	recv := doPost(reqData)
	txNum := hexstr2dec(recv.Result)
	log.Debug(fmt.Sprintf("Block %-6d has %-6d TXs.", blockNum, txNum))
	return txNum
}

// sendTXs sends a batch of tx, batch size is given by parameter count.
func sendTXs(count int) {
	for index := 0; index < count; index++ {
		reqData := fmt.Sprintf(
			`{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from": "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b","to": "%s","gas": "0x6400","gasPrice": "0x1234","value": "0x%x"}],"id":1}`,
			addressList[index%len(addressList)], index+rand.Intn(count))
		doPost(reqData)
	}
}

// doPost is a tool function used to talk to justitia API.
func doPost(reqData string) *RPCResponse {
	endpoint := randomEndpoint()
	request, err := http.NewRequest("POST", endpoint, strings.NewReader(reqData))
	if err != nil {
		log.Error("New request error, please check.")
		log.Error("[POST " + endpoint + "] " + reqData)
		log.Error(fmt.Sprintf("%v", err))
		os.Exit(2)
	}
	response, err := client.Do(request)
	if err != nil {
		log.Error("Send request error, please check.")
		log.Error("[POST " + endpoint + "] " + reqData)
		log.Error(fmt.Sprintf("%v", err))
		os.Exit(2)
	}
	defer response.Body.Close()
	blob, _ := ioutil.ReadAll(response.Body)
	recv := new(RPCResponse)
	json.Unmarshal(blob, recv)
	if recv.Error != nil {
		log.Error("Get response, but has error.")
		log.Error("[POST " + endpoint + "] " + reqData)
		log.Error(fmt.Sprintf("%v", err))
		os.Exit(2)
	}
	return recv
}

// hexstr2dec converts hexadecimal string to decimal integer(int64).
func hexstr2dec(hex string) int64 {
	var str string
	if hex[0:2] == "0x" {
		str = hex[2:]
	} else {
		str = hex
	}
	dec, err := strconv.ParseInt(str, 16, 64)
	if err != nil {
		panic(err)
	}
	return dec
}

// statisticBetweenBlock calculates the tx / second, and blocks / second based
// off of the number the transactions and number of blocks that occurred from
// the start block, to the end time.
func statisticBetweenBlock(beginningHeight int64, endingHeight int64) {

	log.Info(fmt.Sprintf("Beginning block height: %d", beginningHeight))
	log.Info(fmt.Sprintf("Ending block height:    %d", endingHeight))

	timeS := timestampOfBlockByNumber(beginningHeight)
	timeE := timestampOfBlockByNumber(endingHeight)

	log.Debug("Calculating...")
	var totalTxNum int64
	for bn := beginningHeight + 1; bn <= endingHeight; bn++ {
		totalTxNum += txNumOfBlock(bn)
	}

	blockPerSecond := float64(endingHeight-beginningHeight) / float64(timeE-timeS)
	txPerSecond := float64(totalTxNum) / float64(timeE-timeS)

	log.Info(fmt.Sprintf("Bench test result: %0.2f tx/sec, %0.2f block/sec.", txPerSecond, blockPerSecond))
}

// addressList contains addresses used to send tx.
var addressList = []string{
	"0x8be461ea3c27b698a31515a98b8fa339b4bea51a", "0x7a445eaf276834d9aaeda583f46d6b505489923e",
	"0x3be86cf6b79472aa0ad787ec410e08b877e52feb", "0x926794f9785ed0ffe92364ee796f2234998f6f20",
	"0x1d1602e497f7a6d13a4e846ea469a1bfa24ecb13", "0x57924a847e363a49c757792aa2f30f46fa922370",
	"0x1f851d4d373e3e4d93bc1f26718b3ea0e5d3b1f1", "0x01c91a1b352a2903bc8378e5f645c9bc8685029e",
	"0x87f029b41ea019dfbabf17bb579870c3e87faf8a", "0xe09beb1c39b6b50090104fe8cea31a2a9be21739",
	"0xe94ca30fbce78cfee5ef4c6dfa7026cc2017f32a", "0x8450dc7a6afe0e85a54bae972485894ec106703f",
	"0x634b57a395fd4653e7d2cc88ec87c937097305f6", "0xbb24008407076c04d07e8ec244de31b2ef72ca34",
	"0x56eba472d72054ae08937bb7067221f6cab02681", "0xc308768314371211d0cfaeccbc4baa85cc59245e",
	"0xae256b300b1cdac1f868d68575b267b040c90651", "0x7b867ef735257b8fe849ee4f9824fce1c2db88dc",
	"0xc6289a8a63486261f724afd703aa9b1c17c3e077", "0x5f2463638694451f5c673b54469e128b55eb7e9d",
	"0x33edd9354a2653bf40cd753bee238696b27fd519", "0xd42ed9af809230ae32d5ac8a2bd3042f0acb02c6",
	"0x688ce2b649432176cdd9904f7f49670af5445fe1", "0xcfbd1d344c42a4bae76ed689cbd3e57752951810",
	"0x41d34e475eb7dc5894c8bcbc3183829854ef6a76", "0x5b0a7efa42941670b96a2b49cfba44c48dec9374",
	"0x6e859e68c935140591b2fd3a5bb1e0dfcd742fcb", "0xa6fa88efb394e92c676e5b4e87eb490cc31df529",
	"0xe87add574081a25af5600272943746be9587bc48", "0x1e817e1f620597be6609f7d3810280adaa5fda4a",
	"0x1dc94c828126d4834a48ab21884e974e032d07f3", "0xc25a4cf60db650dcfffbe8e9c7f3045327162948",
	"0xe730e71ac2069f75465f4fd71c5d50d07cfc5fc5", "0x87e152f775e91b1cb7bca050677022c49d4f4f83",
	"0x4e57bf3470bf8c214872526269007fcfd4c92d6b", "0x416a8c0045951f3e77d5118368f9341527837906",
	"0xb992ea768f51512657fee52bb91587144cff8b98", "0x3f5b43e961af464fb3ad94ed00c94dd27e2a47e6",
	"0x7dfcd674fd99ada151689c630822f86b91f5d0e6", "0x48396fa1fcb03bd43c4d1e02e3f6024e1552aa4c",
	"0x41fd530e3e91645ebaab218db899d47192606109", "0xf45a43207fed2db3af1cc8e7ca416444082c4b25",
	"0xedd8a3d37be17e2b361868ccbadc78f94fee5fe6", "0x89ff11dab13ed2056cfcb22128a0886d4a989834",
	"0xb4e17991a0d715e3bb9b8a42429eda4026d9b054", "0x95e23e97d88e076df2502faa72a3ba8ad3315ed4",
	"0x1a007089523cc763d8e7c8a2f33429b28cdae5d5", "0x28e708710de4b2e51012b203deec6e02b0927018",
	"0xcb35393297d9ce36247a2ca70d6ee30a130ec254", "0xaa40386ff92635b80c141facbcd6ab1b04b27eb0",
}
