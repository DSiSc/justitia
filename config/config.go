package config

import (
	"fmt"
	blockchainc "github.com/DSiSc/blockchain/config"
	"github.com/DSiSc/craft/monitor"
	consensusc "github.com/DSiSc/galaxy/consensus/config"
	participatesc "github.com/DSiSc/galaxy/participates/config"
	rolec "github.com/DSiSc/galaxy/role/config"
	"github.com/DSiSc/justitia/tools"
	"github.com/DSiSc/txpool"
	"github.com/DSiSc/validator/tools/account"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

const (
	CONFIG_PREFIX = "justitia"
	// algorithm setting
	HASH_ALGORITHM = "general.hashAlgorithm"
	// txpool setting
	TXPOOL_SLOTS  = "general.txpool.globalslots"
	MAX_TXS_BLOCK = "general.txpool.txsPerBlock"
	// consensus policy setting
	CONSENSUS_POLICY    = "general.consensus.policy"
	PARTICIPATES_POLICY = "general.participates.policy"
	PARTICIPATES_NUMBER = "general.participates.participates"
	ROLE_POLICY         = "general.role.policy"
	// node info
	NODE_ADDRESS = "general.node.address"
	// block chain
	BLOCK_CHAIN_PLUGIN     = "general.blockchain.plugin"
	BLOCK_CHAIN_STATE_PATH = "general.blockchain.statePath"
	BLOCK_CHAIN_DATA_PATH  = "general.blockchain.dataPath"
	// api gateway
	API_GATEWAY_TCP_ADDR = "general.apigateway"
	// Default parameter for solo block producer
	SOLO_TEST_BLOCK_PRODUCER_INTERVAL = "general.soloModeBlockProducedInterval"

	// prometheus
	PROMETHEUS_ENABLED  = "monitor.prometheus.enabled"
	PROMETHEUS_PORT     = "monitor.prometheus.port"
	PROMETHEUS_MAX_CONN = "monitor.prometheus.maxOpenConnections"

	// Expvar
	EXPVAR_ENABLED = "monitor.expvar.enabled"
	EXPVAR_PORT    = "monitor.expvar.port"
	EXPVAR_PATH    = "monitor.expvar.path"
)

type AlgorithmConfig struct {
	//hash algorithm
	HashAlgorithm string
	//signature algorithm
	SignAlgorithm string
}

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
	// Block Produce Interval
	BlockInterval uint8
	//algorithm config
	AlgorithmConf AlgorithmConfig

	// prometheus
	PrometheusConf monitor.PrometheusConfig
	// expvar
	ExpvarConf monitor.ExpvarConfig
}

type Config struct {
	filePath string
	maps     map[string]interface{}
}

func LoadConfig() (config *viper.Viper) {
	config = viper.New()

	// for environment variables
	config.SetEnvPrefix(CONFIG_PREFIX)
	config.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)

	config.SetConfigName("justitia")
	homePath, _ := tools.Home()
	config.AddConfigPath(fmt.Sprintf("%s/.justitia", homePath))
	// Path to look for the config file in based on GOPATH
	goPath := os.Getenv("GOPATH")
	for _, p := range filepath.SplitList(goPath) {
		config.AddConfigPath(filepath.Join(p, "src/github.com/DSiSc/justitia/config"))
	}

	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("error reading plugin config: %s", err))
	}
	return
}

func NewNodeConfig() NodeConfig {
	config := LoadConfig()
	algorithmConf := GetAlgorithmConf(config)
	nodeAccount := GetNodeAccount(config)
	apiGatewayTcpAddr := GetApiGatewayTcpAddr(config)
	txPoolConf := NewTxPoolConf(config)
	participatesConf := NewParticipateConf(config)
	roleConf := NewRoleConf(config)
	consensusConf := NewConsensusConf(config)
	blockChainConf := NewBlockChainConf(config)
	blockIntervalTime := GetBlockProducerInterval(config)
	prometheusConf := GetPrometheusConf(config)
	expvarConf := GetExpvarConf(config)

	return NodeConfig{
		Account:          nodeAccount,
		ApiGatewayAddr:   apiGatewayTcpAddr,
		TxPoolConf:       txPoolConf,
		ParticipatesConf: participatesConf,
		RoleConf:         roleConf,
		ConsensusConf:    consensusConf,
		BlockChainConf:   blockChainConf,
		BlockInterval:    blockIntervalTime,
		AlgorithmConf:    algorithmConf,
		PrometheusConf:   prometheusConf,
		ExpvarConf:       expvarConf,
	}

}

func GetAlgorithmConf(config *viper.Viper) AlgorithmConfig {
	policy := config.GetString(HASH_ALGORITHM)
	// TODO get sigure algotihm config
	return AlgorithmConfig{
		HashAlgorithm: policy,
	}
}

func NewTxPoolConf(conf *viper.Viper) txpool.TxPoolConfig {
	slots := conf.GetInt64(TXPOOL_SLOTS)
	txPerBlock := conf.GetInt64(MAX_TXS_BLOCK)
	txPoolConf := txpool.TxPoolConfig{
		GlobalSlots:    uint64(slots),
		MaxTrsPerBlock: uint64(txPerBlock),
	}
	return txPoolConf
}

func NewParticipateConf(conf *viper.Viper) participatesc.ParticipateConfig {
	policy := conf.GetString(PARTICIPATES_POLICY)
	participates := conf.GetInt64(PARTICIPATES_NUMBER)
	participatesConf := participatesc.ParticipateConfig{
		PolicyName: policy,
		Delegates:  uint64(participates),
	}
	return participatesConf
}

func NewRoleConf(conf *viper.Viper) rolec.RoleConfig {
	policy := conf.GetString(ROLE_POLICY)
	roleConf := rolec.RoleConfig{
		PolicyName: policy,
	}
	return roleConf
}

func NewConsensusConf(conf *viper.Viper) consensusc.ConsensusConfig {
	policy := conf.GetString(CONSENSUS_POLICY)
	consensusConf := consensusc.ConsensusConfig{
		PolicyName: policy,
	}
	return consensusConf
}

func NewBlockChainConf(conf *viper.Viper) blockchainc.BlockChainConfig {
	policy := conf.GetString(BLOCK_CHAIN_PLUGIN)
	dataPath := conf.GetString(BLOCK_CHAIN_DATA_PATH)
	statePath := conf.GetString(BLOCK_CHAIN_STATE_PATH)
	blockChainConf := blockchainc.BlockChainConfig{
		PluginName:    policy,
		StateDataPath: statePath,
		BlockDataPath: dataPath,
	}
	return blockChainConf
}

func GetApiGatewayTcpAddr(conf *viper.Viper) string {
	apiGatewayAddr := conf.GetString(API_GATEWAY_TCP_ADDR)
	return apiGatewayAddr
}

func GetNodeAccount(conf *viper.Viper) *account.Account {
	nodeAddr := conf.GetString(NODE_ADDRESS)
	address := tools.HexToAddress(nodeAddr)
	return &account.Account{
		Address: address,
	}
}

func GetBlockProducerInterval(conf *viper.Viper) uint8 {
	blockInterval := conf.GetInt(SOLO_TEST_BLOCK_PRODUCER_INTERVAL)
	return uint8(blockInterval)
}

func GetPrometheusConf(conf *viper.Viper) monitor.PrometheusConfig {
	enabled := conf.GetBool(PROMETHEUS_ENABLED)
	prometheusPort := conf.GetString(PROMETHEUS_PORT)
	prometheusMaxConn := conf.GetInt(PROMETHEUS_MAX_CONN)
	return monitor.PrometheusConfig{
		PrometheusEnabled: enabled,
		PrometheusPort:    prometheusPort,
		PrometheusMaxConn: prometheusMaxConn,
	}
}

func GetExpvarConf(conf *viper.Viper) monitor.ExpvarConfig {
	enabled := conf.GetBool(EXPVAR_ENABLED)
	prometheusPort := conf.GetString(EXPVAR_PORT)
	ExpvarPath := conf.GetString(EXPVAR_PATH)
	return monitor.ExpvarConfig{
		ExpvarEnabled: enabled,
		ExpvarPort:    prometheusPort,
		ExpvarPath:    ExpvarPath,
	}
}
