package config

import (
	"encoding/json"
	consensus_c "github.com/DSiSc/galaxy/consensus/config"
	participates_c "github.com/DSiSc/galaxy/participates/config"
	role_c "github.com/DSiSc/galaxy/role/config"
	ledger_c "github.com/DSiSc/ledger/config"
	producer_c "github.com/DSiSc/producer/config"
	"github.com/DSiSc/txpool"
	"github.com/DSiSc/txpool/common"
	"github.com/DSiSc/txpool/common/log"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var ConfigName = "config.json"
var DefaultDataDir = "./config"

const (
	// txpool setting
	TXPOOL_SLOTS = "txpool.globalSlots"
	// producer setting
	PRODUCER_TIMER             = "timer"
	PRODUCER_POLICY            = "producer.policy"
	PRODUCER_POLICY_TIMER      = "producer.timer"
	PRODUCER_POLICY_TIMER_TIME = "producer.timer.time"
	// consensus policy setting
	CONSENSUS_POLICY    = "consensus.policy"
	PARTICIPATES_POLICY = "participates.policy"
	ROLE_POLICY         = "role.policy"
	// ledger store setting
	DB_STORE_PLUGIN = "block.plugin"
	DB_STORE_PATH   = "block.path"
)

type NodeConfig struct {
	// default
	Account common.Address
	// txpool
	TxPoolConf txpool.TxPoolConfig
	// producer
	ProducerConf producer_c.ProducerConfig
	// participates
	ParticipatesConf participates_c.ParticipateConfig
	// role
	RoleConf role_c.RoleConfig
	// consensus
	ConsensusConf consensus_c.ConsensusConfig
	// ledger
	LedgerConf ledger_c.LedgerConfig
}

type Config struct {
	filePath string
	maps     map[string]interface{}
}

func New(path string) Config {
	_, file, _, _ := runtime.Caller(1)
	keyString := "/github.com/DSiSc/justitia/"
	index := strings.LastIndex(file, keyString)
	confAbsPath := strings.Join([]string{file[:index+len(keyString)], "config/config.json"}, "")
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
	var temp common.Address
	conf := New(ConfigName)
	txPoolConf := conf.NewTxPoolConf()
	producerConf := conf.NewProducerConf()
	participatesConf := conf.NewParticipateConf()
	roleConf := conf.NewRoleConf()
	consensusConf := conf.NewConsensusConf()
	ledgerConf := conf.NewLedgerConf()

	// TODO: get account and globalSlots from genesis.json
	return NodeConfig{
		Account:          temp,
		TxPoolConf:       txPoolConf,
		ProducerConf:     producerConf,
		ParticipatesConf: participatesConf,
		RoleConf:         roleConf,
		ConsensusConf:    consensusConf,
		LedgerConf:       ledgerConf,
	}
}

func (self *Config) NewTxPoolConf() txpool.TxPoolConfig {
	slots, err := strconv.ParseUint(self.GetConfigItem(TXPOOL_SLOTS).(string), 10, 64)
	if err != nil {
		log.Error("Get slots failed.")
	}
	txPoolConf := txpool.TxPoolConfig{
		GlobalSlots: slots,
	}
	return txPoolConf
}

func (self *Config) NewProducerConf() producer_c.ProducerConfig {
	policy := self.GetConfigItem(PRODUCER_POLICY).(string)
	time, err := strconv.ParseUint(self.GetConfigItem(PRODUCER_POLICY_TIMER_TIME).(string), 10, 64)
	if err != nil {
		log.Error("Get time for producer failed.")
	}
	producerConf := producer_c.ProducerConfig{
		PolicyName: policy,
		PolicyContext: producer_c.ProducerPolicy{
			Timer: time,
			Num:   1,
		},
	}
	return producerConf
}

func (self *Config) NewParticipateConf() participates_c.ParticipateConfig {
	policy := self.GetConfigItem(PARTICIPATES_POLICY).(string)
	participatesConf := participates_c.ParticipateConfig{
		PolicyName: policy,
	}
	return participatesConf
}

func (self *Config) NewRoleConf() role_c.RoleConfig {
	policy := self.GetConfigItem(ROLE_POLICY).(string)
	roleConf := role_c.RoleConfig{
		PolicyName: policy,
	}
	return roleConf
}

func (self *Config) NewConsensusConf() consensus_c.ConsensusConfig {
	policy := self.GetConfigItem(CONSENSUS_POLICY).(string)
	consensusConf := consensus_c.ConsensusConfig{
		PolicyName: policy,
	}
	return consensusConf
}

func (self *Config) NewLedgerConf() ledger_c.LedgerConfig {
	policy := self.GetConfigItem(DB_STORE_PLUGIN).(string)
	dataPath := self.GetConfigItem(DB_STORE_PATH).(string)
	ledgerConf := ledger_c.LedgerConfig{
		PluginName: policy,
		DataPath:   dataPath,
	}
	return ledgerConf
}
