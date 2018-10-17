package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var address = []string{
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

func benchTest(url string, count uint64) {
	client := &http.Client{}
	var index uint64
	for index = uint64(0); index < count; index++ {
		fmt.Printf("################### result of tx %d is: ###################\n", index)
		mm := fmt.Sprintf("0x%x", index/uint64(len(address))+1)
		// fmt.Println(mm)
		payload := fmt.Sprintf(
			`{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from": "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b","to": "%s","gas": "0x6400","gasPrice": "%s","value": "0x1"}],"id":1}`,
			address[index%uint64(len(address))], mm)
		reqest, err := http.NewRequest("POST", url, strings.NewReader(payload))
		if err != nil {
			panic(err)
		}
		response, err := client.Do(reqest)
		if err != nil {
			fmt.Println("Do request error, please confirm.")
			continue
		}
		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(body))
	}
}

func main() {
	ip := flag.String("ip", "127.0.0.1", "http listen address.")
	port := flag.String("port", "47768", "http listen port.")
	span := flag.Int64("span", 50, "Random number of transactions.")
	flag.Parse()
	url := fmt.Sprintf("http://%s:%s", *ip, *port)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fmt.Println(url)
	for {
		random := r.Int63n(*span)
		fmt.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> Send %d Transactions.\n", random)
		benchTest(url, uint64(random))
		time.Sleep(2 * time.Second)
	}
}
