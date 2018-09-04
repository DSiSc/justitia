package config

import (
	"encoding/json"
	blockchainc "github.com/DSiSc/blockchain/config"
	consensusc "github.com/DSiSc/galaxy/consensus/config"
	participatesc "github.com/DSiSc/galaxy/participates/config"
	rolec "github.com/DSiSc/galaxy/role/config"
	"github.com/DSiSc/justitia/tools"
	"github.com/DSiSc/txpool"
	"github.com/DSiSc/txpool/log"
	"github.com/DSiSc/validator/tools/account"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var ConfigName = "config.json"
var DefaultDataDir = "./config"

const (
	// json file relative path
	CONFIG_DIR = "config/"
	// txpool setting
	TXPOOL_SLOTS  = "txpool.globalSlots"
	MAX_TXS_BLOCK = "txpool.txsPerBlock"
	// consensus policy setting
	CONSENSUS_POLICY    = "consensus.policy"
	PARTICIPATES_POLICY = "participates.policy"
	ROLE_POLICY         = "role.policy"
	// node info
	NODE_ADDRESS = "node.address"
	// block chain
	BLOCK_CHAIN_PLUGIN     = "blockchain.plugin"
	BLOCK_CHAIN_STATE_PATH = "blockchain.statePath"
	BLOCK_CHAIN_DATA_PATH  = "blockchain.dataPath"
	// api gateway
	API_GATEWAY_TCP_ADDR = "apigateway.tcpAddr"
)

type NodeConfig struct {
	// default
	Account *account.Account
	// api gateway
	ApiGatewayAddr string
	// txpool
	TxPoolConf txpool.TxPoolConfig
	// participates
	ParticipatesConf participatesc.ParticipateConfig
	// role
	RoleConf rolec.RoleConfig
	// consensus
	ConsensusConf consensusc.ConsensusConfig
	// BlockChainConfig
	BlockChainConf blockchainc.BlockChainConfig
}

type Config struct {
	filePath string
	maps     map[string]interface{}
}

func New(path string) Config {
	_, file, _, _ := runtime.Caller(1)
	keyString := "/github.com/DSiSc/justitia/"
	index := strings.LastIndex(file, keyString)
	relPath := CONFIG_DIR + ConfigName
	confAbsPath := strings.Join([]string{file[:index+len(keyString)], relPath}, "")
	return Config{filePath: confAbsPath}
}

// Read the given json file.
func (config *Config) read() {
	if !filepath.IsAbs(config.filePath) {
		filePath, err := filepath.Abs(config.filePath)
		if err != nil {
			panic(err)
		}
		config.filePath = filePath
	}

	bts, err := ioutil.ReadFile(config.filePath)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bts, &config.maps)

	if err != nil {
		panic(err)
	}
}

// If we want to get item in a stucture, which like this:
//{
//	"classs": {
//		"student":{
//			"name": "john"
//         }
//     }
//}
// { class: {}}
// You can get it by call Get("class.student.name")
func (config *Config) GetConfigItem(name string) interface{} {
	if config.maps == nil {
		config.read()
	}

	if config.maps == nil {
		return nil
	}

	keys := strings.Split(name, ".")
	length := len(keys)
	if length == 1 {
		return config.maps[name]
	}

	var ret interface{}
	for i := 0; i < length; i++ {
		if i == 0 {
			ret = config.maps[keys[i]]
			if ret == nil {
				return nil
			}
		} else {
			if m, ok := ret.(map[string]interface{}); ok {
				ret = m[keys[i]]
			} else {
				if length == i-1 {
					return ret
				}
				return nil
			}
		}
	}
	return ret
}

func NewNodeConfig() NodeConfig {
	conf := New(ConfigName)
	nodeAccount := conf.GetNodeAccount()
	apiGatewayTcpAddr := conf.GetApiGatewayTcpAddr()
	txPoolConf := conf.NewTxPoolConf()
	participatesConf := conf.NewParticipateConf()
	roleConf := conf.NewRoleConf()
	consensusConf := conf.NewConsensusConf()
	blockChainConf := conf.NewBlockChainConf()

	return NodeConfig{
		Account:          nodeAccount,
		ApiGatewayAddr:   apiGatewayTcpAddr,
		TxPoolConf:       txPoolConf,
		ParticipatesConf: participatesConf,
		RoleConf:         roleConf,
		ConsensusConf:    consensusConf,
		BlockChainConf:   blockChainConf,
	}
}

func (self *Config) NewTxPoolConf() txpool.TxPoolConfig {
	slots, err := strconv.ParseUint(self.GetConfigItem(TXPOOL_SLOTS).(string), 10, 64)
	if err != nil {
		log.Error("Get slots failed.")
		slots = 0
	}
	txs, err1 := strconv.ParseUint(self.GetConfigItem(MAX_TXS_BLOCK).(string), 10, 64)
	if nil != err1 {
		log.Error("Get slots failed.")
		txs = 0
	}
	txPoolConf := txpool.TxPoolConfig{
		GlobalSlots:    slots,
		MaxTrxPerBlock: txs,
	}
	return txPoolConf
}

func (self *Config) NewParticipateConf() participatesc.ParticipateConfig {
	policy := self.GetConfigItem(PARTICIPATES_POLICY).(string)
	participatesConf := participatesc.ParticipateConfig{
		PolicyName: policy,
	}
	return participatesConf
}

func (self *Config) NewRoleConf() rolec.RoleConfig {
	policy := self.GetConfigItem(ROLE_POLICY).(string)
	roleConf := rolec.RoleConfig{
		PolicyName: policy,
	}
	return roleConf
}

func (self *Config) NewConsensusConf() consensusc.ConsensusConfig {
	policy := self.GetConfigItem(CONSENSUS_POLICY).(string)
	consensusConf := consensusc.ConsensusConfig{
		PolicyName: policy,
	}
	return consensusConf
}

func (self *Config) NewBlockChainConf() blockchainc.BlockChainConfig {
	policy := self.GetConfigItem(BLOCK_CHAIN_PLUGIN).(string)
	dataPath := self.GetConfigItem(BLOCK_CHAIN_DATA_PATH).(string)
	statePath := self.GetConfigItem(BLOCK_CHAIN_STATE_PATH).(string)
	blockChainConf := blockchainc.BlockChainConfig{
		PluginName:    policy,
		StateDataPath: statePath,
		BlockDataPath: dataPath,
	}
	return blockChainConf
}

func (self *Config) GetApiGatewayTcpAddr() string {
	apiGatewayAddr := self.GetConfigItem(API_GATEWAY_TCP_ADDR).(string)
	return apiGatewayAddr
}

func (self *Config) GetNodeAccount() *account.Account {
	nodeAddr := self.GetConfigItem(NODE_ADDRESS).(string)
	address := tools.HexToAddress(nodeAddr)
	return &account.Account{
		Address: address,
	}
}
