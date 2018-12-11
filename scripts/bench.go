package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/DSiSc/craft/log"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
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

type timeType uint8

const (
	NullTime = timeType(iota)
	Year
	Month
	Day
	Hour
	Minute
	Second
)

const (
	SecondBase     = float64(1)
	MinuteToSecond = 60 * SecondBase
	HourToSecond   = 60 * MinuteToSecond
)

const (
	BlockAssembleTxsKeyWords = "assemble success with"
	DefaultBlockStartIndex   = 1
	DefaultBlockEndIndex     = 1
)

type LogStruct struct {
	Level     string `json:"level"`
	HostName  string `json:"host"`
	Timestamp string `json:"time"`
	Caller    string `json:"caller"`
	Message   string `json:"message"`
}

var (
	client   = &http.Client{}
	endpoint string
)

// main process goes here.
func main() {
	var durationInt, txsRate, blockStart, blockEnd int
	var verbose, random, showHelp, statistic bool
	var timerWG sync.WaitGroup
	var filePath string

	//////////////////////////////////
	// flagSet handles command flags.
	//////////////////////////////////
	flagSet := flag.NewFlagSet("jt-bench", flag.ExitOnError)
	flagSet.IntVar(&durationInt, "t", 30, "Exit after the specified amount of time in seconds")
	flagSet.IntVar(&txsRate, "r", 200, "Txs per second to send in a connection")
	flagSet.BoolVar(&verbose, "v", false, "Verbose output")
	flagSet.BoolVar(&showHelp, "h", false, "Display help")
	flagSet.BoolVar(&random, "random", false, "Random number of tx")
	flagSet.BoolVar(&statistic, "x", false, "The switch of statistic and bench.")
	flagSet.IntVar(&blockStart, "s", 1, "The number of starting statistics.")
	flagSet.IntVar(&blockEnd, "e", 1, "The number of ending statistics.")
	flagSet.StringVar(&filePath, "f", "/var/log/justitia/justitia.log", "The absolute path of statistic file.")
	flagSet.Usage = func() {
		fmt.Println(`Justitia blockchain benchmarking tool.

Usage:
	go run bench.go [-t 30] [-r 200] [-v] [-x] [-s] [-e] [endpoint]

Examples:
    There are two ways to use this tools.

    First is to bench, simulate send 200 txs/s in 30 second
    It will output starting, ending block number and tps, bps at the same time.

	go run bench.go -t 30 -r 200 http://127.0.0.1:47768

    Second is statistic output of the bench result, it will output log info
    when making block, and tps, bps at the same time.

	go run bench.go -x -s 117 -e 160

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
		endpoint = "http://127.0.0.1:47768"
	} else {
		endpoint = flagSet.Arg(0)
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

	if statistic {
		resultStatistic(filePath, blockStart, blockEnd)
		return
	}

	//////////////////////////////////
	// send tx at a given rate
	//////////////////////////////////
	var ticker, totalTimer *time.Ticker
	timerWG.Add(2)
	go func() {
		log.Info(fmt.Sprintf("Sending TX at rate of %d tx/sec ...", txsRate))
		ticker = time.NewTicker(1 * time.Second)
		sendTXs(txsRate)
		for i := 0; i <= durationInt; i++ {
			<-ticker.C
			if random {
				sendTXs(rand.Intn(txsRate * 2))
			} else {
				sendTXs(txsRate)
			}
		}
		timerWG.Done()
	}()

	//////////////////////////////////////////////////////////////
	// accumulate tx number of blocks generated in bench period.
	//////////////////////////////////////////////////////////////
	go func() {
		// wait a moment for tx to be stored into block.
		time.Sleep(time.Millisecond * 1000)

		totalTimer = time.NewTicker(time.Second * time.Duration(durationInt))

		beginHeight := latestBlockNumber()
		log.Info(fmt.Sprintf("Record beginnng block height: %d", beginHeight))
		<-totalTimer.C
		endHeight := latestBlockNumber()
		log.Info(fmt.Sprintf("Record ending block height: %d", endHeight))

		//////////////////////////////////////////////////////////////////////////////
		// calculateStatistics calculates the tx / second, and blocks / second based
		// off of the number the transactions and number of blocks that occurred from
		// the start block, to the end time.
		//////////////////////////////////////////////////////////////////////////////
		log.Info("Calculating...")
		var totalTxNum int64
		for bn := beginHeight; bn <= endHeight; bn++ {
			totalTxNum += txNumOfBlock(bn)
		}

		blockPerSecond := float64(endHeight-beginHeight) / float64(durationInt)
		txPerSecond := float64(totalTxNum) / float64(durationInt)

		log.Info(fmt.Sprintf("Bench test result: %0.2f tx/sec, %0.2f block/sec.", txPerSecond, blockPerSecond))

		timerWG.Done()
	}()

	timerWG.Wait()
}

// latestBlockNumber fetches current block number.
func latestBlockNumber() int64 {
	reqData := `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":8}`
	recv := doPost(reqData)
	return hexstr2dec(recv.Result)
}

// txNumOfBlock gets tx number of given block number.
func txNumOfBlock(blockNum int64) int64 {
	reqData := fmt.Sprintf(`{"jsonrpc": "2.0", "method": "eth_getBlockTransactionCountByNumber", "id": 1, "params": ["0x%x"]}`, blockNum)
	recv := doPost(reqData)
	txNum := hexstr2dec(recv.Result)
	log.Debug(fmt.Sprintf("Block[%d] has %d TXs.", blockNum, txNum))
	return txNum
}

// sendTXs sends a batch of tx, batch size is given by parameter count.
func sendTXs(count int) {
	log.Debug("Sending %d Txs...", count)
	for index := 0; index < count; index++ {
		reqData := fmt.Sprintf(
			`{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from": "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b","to": "%s","gas": "0x6400","gasPrice": "0x1234","value": "0x%x"}],"id":1}`,
			addressList[index%len(addressList)], index)
		doPost(reqData)
	}
}

// doPost is a tool function used to talk to justitia API.
func doPost(reqData string) *RPCResponse {
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

func minusSecond(x float64, y float64, style interface{}) float64 {
	if Hour == style {
		return (x - y) * HourToSecond
	}
	if Minute == style {
		return (x - y) * MinuteToSecond
	}
	if Second == style {
		return (x - y) * SecondBase
	}
	return float64(0)
}

func computeTime(startTime string, endTime string) float64 {
	start := strings.Split(startTime, " ")
	startSplit := strings.Split(start[1], ":")
	startHour, _ := strconv.ParseFloat(startSplit[0], 64)
	startMin, _ := strconv.ParseFloat(startSplit[1], 64)
	startSec, _ := strconv.ParseFloat(startSplit[2], 64)
	end := strings.Split(endTime, " ")
	endSplit := strings.Split(end[1], ":")
	endHour, _ := strconv.ParseFloat(endSplit[0], 64)
	endMin, _ := strconv.ParseFloat(endSplit[1], 64)
	endSec, _ := strconv.ParseFloat(endSplit[2], 64)
	diffHour := minusSecond(endHour, startHour, Hour)
	diffMinute := minusSecond(endMin, startMin, Minute)
	diffSecond := minusSecond(endSec, startSec, Second)
	totalTime := diffHour + diffMinute + diffSecond
	fmt.Printf("diffHour: %v.\ndiffMinute: %v.\ndiffSecond: %v.\ntotal time %v.\n", diffHour, diffMinute, diffSecond, totalTime)
	return totalTime
}

func resultStatistic(filePath string, blockStart int, blockEnd int) {
	fmt.Printf("start %d, end %d.\n", blockStart, blockEnd)
	var totalTxs = uint64(0)
	var startTime, endTime string
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("Open file error!", err)
		return
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	var success = false
	var startIndex = blockStart
	if DefaultBlockEndIndex == blockEnd {
		for {
			line, err := buf.ReadString('\n')
			line = strings.TrimSpace(line)
			var keyString = fmt.Sprintf("Block %d %s", startIndex, BlockAssembleTxsKeyWords)
			if strings.Contains(line, keyString) {
				success = true
				var logLine LogStruct
				if err := json.Unmarshal([]byte(line), &logLine); err == nil {
					if 0 == totalTxs {
						startTime = logLine.Timestamp
					}
					reg := regexp.MustCompile(`(?U)\b.+\b`)
					txString := reg.FindAllString(logLine.Message, -1)
					tx, _ := strconv.Atoi(txString[10])
					totalTxs = totalTxs + uint64(tx)
					fmt.Printf("block: %d, txs: %d, time %s.\n", startIndex, tx, logLine.Timestamp)
					if 0 == tx {
						endTime = logLine.Timestamp
						fmt.Printf("Total txs %d.\nstart time %s.\nend time %s.\n", totalTxs, startTime, endTime)
						break
					}
				} else {
					fmt.Println(err)
				}
				startIndex = startIndex + 1
			}
			if err != nil {
				if err == io.EOF {
					fmt.Println("File read ok!")
					break
				} else {
					fmt.Println("Read file error!", err)
					return
				}
			}
		}
	} else {
		for {
			if startIndex < blockEnd {
				line, err := buf.ReadString('\n')
				line = strings.TrimSpace(line)
				var keyString = fmt.Sprintf("Block %d %s", startIndex, BlockAssembleTxsKeyWords)
				if strings.Contains(line, keyString) {
					success = true
					var logLine LogStruct
					if err := json.Unmarshal([]byte(line), &logLine); err == nil {
						if 0 == totalTxs {
							startTime = logLine.Timestamp
						}
						reg := regexp.MustCompile(`(?U)\b.+\b`)
						txString := reg.FindAllString(logLine.Message, -1)
						tx, _ := strconv.Atoi(txString[10])
						totalTxs = totalTxs + uint64(tx)
						fmt.Printf("block: %d, txs: %d, time %s.\n", startIndex, tx, logLine.Timestamp)
						endTime = logLine.Timestamp
					} else {
						fmt.Println(err)
					}
					startIndex = startIndex + 1
				}
				if err != nil {
					if err == io.EOF {
						fmt.Println("File read ok!")
						break
					} else {
						fmt.Println("Read file error!", err)
						return
					}
				}
			} else {
				break
			}
		}
	}
	if success {
		totalTime := computeTime(startTime, endTime)
		rate := float64(totalTxs) / totalTime
		bps := float64(blockEnd-blockStart) / totalTime
		fmt.Printf("system average TPS: %v txs/s.\n", rate)
		fmt.Printf("system average BPS: %v b/s.\n", bps)
		return
	}
	fmt.Println("no record found, please confirm.")
}
