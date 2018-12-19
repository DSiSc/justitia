package config

import (
	"fmt"
	blockchainc "github.com/DSiSc/blockchain/config"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/monitor"
	consensusc "github.com/DSiSc/galaxy/consensus/config"
	participatesc "github.com/DSiSc/galaxy/participates/config"
	rolec "github.com/DSiSc/galaxy/role/config"
	"github.com/DSiSc/justitia/common"
	"github.com/DSiSc/justitia/tools"
	p2pConf "github.com/DSiSc/p2p/config"
	"github.com/DSiSc/txpool"
	"github.com/DSiSc/validator/tools/account"
	"github.com/spf13/viper"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const (
	// config file prefix
	ConfigPrefix = "justitia"
	// node type
	NodeType = "general.nodeType"
	// algorithm setting
	HashAlgorithm = "general.hashAlgorithm"
	// txpool setting
	TxpoolSlots = "general.txpool.globalslots"
	MaxTxBlock  = "general.txpool.txsPerBlock"
	// consensus policy setting
	ConsensusPolicy    = "general.consensus.policy"
	ConsensusTimeout   = "general.consensus.timeout"
	ParticipatesPolicy = "general.participates.policy"
	ParticipatesNumber = "general.participates.participates"
	RolePolicy         = "general.role.policy"
	// node info
	NodeAddress = "general.node.address"
	NodeId      = "general.node.id"
	NodeUrl     = "general.node.url"
	// block chain
	BlockChainPlugin    = "general.blockchain.plugin"
	BlockChainStatePath = "general.blockchain.statePath"
	BlockChainDataPath  = "general.blockchain.dataPath"
	// api gateway
	ApiGatewayAddr = "general.apigateway"
	// Default parameter for solo block producer
	SoloModeProducerInterval = "general.soloModeBlockProducedInterval"

	//P2P Setting
	// block syncer p2p config
	BlockSyncerP2p                = "block_syncer_p2p"
	P2pBlockSyncerAddrBook        = "general.p2p.blockSyncer.AddrBookFilePath"
	P2pBlockSyncerListenAddr      = "general.p2p.blockSyncer.ListenAddress"
	P2pBlockSyncerMaxOut          = "general.p2p.blockSyncer.MaxConnOutBound"
	P2pBlockSyncerMaxIn           = "general.p2p.blockSyncer.MaxConnInBound"
	P2pBlockSyncerPersistentPeers = "general.p2p.blockSyncer.PersistentPeers"

	// block p2p config
	BlockP2p                = "block_p2p"
	P2pBlockAddrBook        = "general.p2p.block.AddrBookFilePath"
	P2pBlockListenAddr      = "general.p2p.block.ListenAddress"
	P2pBlockMaxOut          = "general.p2p.block.MaxConnOutBound"
	P2pBlockMaxIn           = "general.p2p.block.MaxConnInBound"
	P2pBlockPersistentPeers = "general.p2p.block.PersistentPeers"

	// tx p2p config
	TxP2p                = "tx_p2p"
	P2pTxAddrBook        = "general.p2p.tx.AddrBookFilePath"
	P2pTxListenAddr      = "general.p2p.tx.ListenAddress"
	P2pTxMaxOut          = "general.p2p.tx.MaxConnOutBound"
	P2pTxMaxIn           = "general.p2p.tx.MaxConnInBound"
	P2pTxPersistentPeers = "general.p2p.tx.PersistentPeers"

	// prometheus
	PrometheusEnabled = "monitor.prometheus.enabled"
	PrometheusPort    = "monitor.prometheus.port"
	PrometheusMaxConn = "monitor.prometheus.maxOpenConnections"

	// Expvar
	ExpvarEnabled = "monitor.expvar.enabled"
	ExpvarPort    = "monitor.expvar.port"
	ExpvarPath    = "monitor.expvar.path"

	// pprof
	PprofEnabled = "monitor.pprof.enabled"
	PprofPort    = "monitor.pprof.port"

	// Log Setting
	LogTimeFieldFormat = "logging.timeFieldFormat"
	LogConsoleEnabled  = "logging.console.enabled"
	LogConsoleLevel    = "logging.console.level"
	LogConsoleFormat   = "logging.console.format"
	LogConsoleCaller   = "logging.console.caller"
	LogConsoleHostname = "logging.console.hostname"
	LogFileEnabled     = "logging.file.enabled"
	LogFilePath        = "logging.file.path"
	LogFileLevel       = "logging.file.level"
	LogFileFormat      = "logging.file.format"
	LogFileCaller      = "logging.file.caller"
	LogFileHostname    = "logging.file.hostname"
)

type AlgorithmConfig struct {
	//hash algorithm
	HashAlgorithm string
	//signature algorithm
	SignAlgorithm string
}

type NodeConfig struct {
	NodeType common.NodeType
	// default
	Account account.Account
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
	BlockInterval int64
	//algorithm config
	AlgorithmConf AlgorithmConfig

	// prometheus
	PrometheusConf monitor.PrometheusConfig
	// expvar
	ExpvarConf monitor.ExpvarConfig
	// pprof
	PprofConf monitor.PprofConfig
	// log setting
	Logger log.Config
	//P2P config
	P2PConf map[string]*p2pConf.P2PConfig
}

type Config struct {
	filePath string
	maps     map[string]interface{}
}

func LoadConfig() (config *viper.Viper) {
	config = viper.New()
	// for environment variables
	config.SetEnvPrefix(ConfigPrefix)
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
	nodeType := getNodeType(config)
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
	pprofConf := GetPprofConf(config)
	logConf := GetLogSetting(config)
	p2pConfs := GetP2PConf(config)

	return NodeConfig{
		Account:          nodeAccount,
		NodeType:         nodeType,
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
		PprofConf:        pprofConf,
		Logger:           logConf,
		P2PConf:          p2pConfs,
	}
}

func GetAlgorithmConf(config *viper.Viper) AlgorithmConfig {
	policy := config.GetString(HashAlgorithm)
	// TODO get sigure algotihm config
	return AlgorithmConfig{
		HashAlgorithm: policy,
	}
}

func NewTxPoolConf(conf *viper.Viper) txpool.TxPoolConfig {
	slots := conf.GetInt64(TxpoolSlots)
	txPerBlock := conf.GetInt64(MaxTxBlock)
	txPoolConf := txpool.TxPoolConfig{
		GlobalSlots:    uint64(slots),
		MaxTrsPerBlock: uint64(txPerBlock),
	}
	return txPoolConf
}

func NewParticipateConf(conf *viper.Viper) participatesc.ParticipateConfig {
	policy := conf.GetString(ParticipatesPolicy)
	participates := conf.GetInt64(ParticipatesNumber)
	participatesConf := participatesc.ParticipateConfig{
		PolicyName: policy,
		Delegates:  uint64(participates),
	}
	return participatesConf
}

func NewRoleConf(conf *viper.Viper) rolec.RoleConfig {
	policy := conf.GetString(RolePolicy)
	roleConf := rolec.RoleConfig{
		PolicyName: policy,
	}
	return roleConf
}

func NewConsensusConf(conf *viper.Viper) consensusc.ConsensusConfig {
	policy := conf.GetString(ConsensusPolicy)
	timeout := conf.GetInt64(ConsensusTimeout)
	consensusConf := consensusc.ConsensusConfig{
		PolicyName: policy,
		Timeout:    timeout,
	}
	return consensusConf
}

func NewBlockChainConf(conf *viper.Viper) blockchainc.BlockChainConfig {
	policy := conf.GetString(BlockChainPlugin)
	dataPath := conf.GetString(BlockChainDataPath)
	statePath := conf.GetString(BlockChainStatePath)
	blockChainConf := blockchainc.BlockChainConfig{
		PluginName:    policy,
		StateDataPath: statePath,
		BlockDataPath: dataPath,
	}
	return blockChainConf
}

func GetApiGatewayTcpAddr(conf *viper.Viper) string {
	apiGatewayAddr := conf.GetString(ApiGatewayAddr)
	return apiGatewayAddr
}

func GetNodeAccount(conf *viper.Viper) account.Account {
	nodeAddr := conf.GetString(NodeAddress)
	address := tools.HexToAddress(nodeAddr)
	id := conf.GetInt64(NodeId)
	url := conf.GetString(NodeUrl)
	return account.Account{
		Address: address,
		Extension: account.AccountExtension{
			Id:  uint64(id),
			Url: url,
		},
	}
}

func GetBlockProducerInterval(conf *viper.Viper) int64 {
	blockInterval := conf.GetInt64(SoloModeProducerInterval)
	return blockInterval
}

func GetPrometheusConf(conf *viper.Viper) monitor.PrometheusConfig {
	enabled := conf.GetBool(PrometheusEnabled)
	prometheusPort := conf.GetString(PrometheusPort)
	prometheusMaxConn := conf.GetInt(PrometheusMaxConn)
	return monitor.PrometheusConfig{
		PrometheusEnabled: enabled,
		PrometheusPort:    prometheusPort,
		PrometheusMaxConn: prometheusMaxConn,
	}
}

func GetExpvarConf(conf *viper.Viper) monitor.ExpvarConfig {
	enabled := conf.GetBool(ExpvarEnabled)
	prometheusPort := conf.GetString(ExpvarPort)
	ExpvarPath := conf.GetString(ExpvarPath)
	return monitor.ExpvarConfig{
		ExpvarEnabled: enabled,
		ExpvarPort:    prometheusPort,
		ExpvarPath:    ExpvarPath,
	}
}

func GetPprofConf(conf *viper.Viper) monitor.PprofConfig {
	enabled := conf.GetBool(PprofEnabled)
	pprofPort := conf.GetString(PprofPort)
	return monitor.PprofConfig{
		PprofEnabled: enabled,
		PprofPort:    pprofPort,
	}
}

func GetLogSetting(conf *viper.Viper) log.Config {
	logTimestampFormat := conf.GetString(LogTimeFieldFormat)
	logConsoleEnabled := conf.GetBool(LogConsoleEnabled)
	logConsoleLevel := conf.GetInt(LogConsoleLevel)
	logConsoleFormat := conf.GetString(LogTimeFieldFormat)
	logConsoleCaller := conf.GetBool(LogConsoleCaller)
	logConsoleHostname := conf.GetBool(LogConsoleHostname)
	logFileEnabled := conf.GetBool(LogFileEnabled)
	logFilePath := conf.GetString(LogFilePath)
	logFileLevel := conf.GetInt(LogFileLevel)
	logFileFormat := conf.GetString(LogFileFormat)
	logFileCaller := conf.GetBool(LogFileCaller)
	logFileHostname := conf.GetBool(LogFileHostname)

	consoleAppender := &log.Appender{
		LogLevel:     log.Level(logConsoleLevel),
		Output:       os.Stdout,
		Format:       strings.ToUpper(logConsoleFormat),
		ShowCaller:   logConsoleCaller,
		ShowHostname: logConsoleHostname,
	}
	logfile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	fileAppender := &log.Appender{
		LogLevel:     log.Level(logFileLevel),
		Output:       logfile,
		Format:       strings.ToUpper(logFileFormat),
		ShowCaller:   logFileCaller,
		ShowHostname: logFileHostname,
	}
	globalLogConfig := log.Config{
		Enabled:  logConsoleEnabled && logFileEnabled,
		Provider: log.GetGlobalConfig().Provider,
		//Provider:        log.Zerolog,
		GlobalLogLevel:  log.Level(uint8(math.Max(float64(logConsoleLevel), float64(logFileLevel)))),
		TimeFieldFormat: logTimestampFormat,
		Appenders:       map[string]*log.Appender{"consolelog": consoleAppender, "filelog": fileAppender},
		OutputFlags:     log.GetOutputFlags(),
		//OutputFlags: &log.OutputFlags{
		//	TimestampFieldName: "time",
		//	LevelFieldName:     "level",
		//	MessageFieldName:   "message",
		//	ErrorFieldName:     "error",
		//	CallerFieldName:    "caller",
		//	HostnameFieldName:  "host",
		//},
	}
	return globalLogConfig
}

func GetP2PConf(conf *viper.Viper) map[string]*p2pConf.P2PConfig {
	p2pConfig := make(map[string]*p2pConf.P2PConfig)
	p2pConfig[BlockSyncerP2p] = getBlockSyncerP2PConf(conf)
	p2pConfig[BlockP2p] = getBlockP2PConf(conf)
	p2pConfig[TxP2p] = getTxP2PConf(conf)
	return p2pConfig
}

func getBlockSyncerP2PConf(conf *viper.Viper) *p2pConf.P2PConfig {
	addrFile := conf.GetString(P2pBlockSyncerAddrBook)
	listenAddr := conf.GetString(P2pBlockListenAddr)
	maxOut := conf.GetInt(P2pBlockMaxIn)
	maxIn := conf.GetInt(P2pBlockMaxOut)
	persistentPeers := conf.GetString(P2pBlockSyncerPersistentPeers)
	return &p2pConf.P2PConfig{
		AddrBookFilePath: addrFile,
		ListenAddress:    listenAddr,
		MaxConnOutBound:  maxOut,
		MaxConnInBound:   maxIn,
		PersistentPeers:  persistentPeers,
	}
}

func getBlockP2PConf(conf *viper.Viper) *p2pConf.P2PConfig {
	addrFile := conf.GetString(P2pBlockAddrBook)
	listenAddr := conf.GetString(P2pBlockListenAddr)
	maxOut := conf.GetInt(P2pBlockMaxOut)
	maxIn := conf.GetInt(P2pBlockMaxIn)
	persistentPeers := conf.GetString(P2pBlockPersistentPeers)
	return &p2pConf.P2PConfig{
		AddrBookFilePath: addrFile,
		ListenAddress:    listenAddr,
		MaxConnOutBound:  maxOut,
		MaxConnInBound:   maxIn,
		PersistentPeers:  persistentPeers,
	}
}

func getTxP2PConf(conf *viper.Viper) *p2pConf.P2PConfig {
	addrFile := conf.GetString(P2pTxAddrBook)
	listenAddr := conf.GetString(P2pTxListenAddr)
	maxOut := conf.GetInt(P2pTxMaxIn)
	maxIn := conf.GetInt(P2pTxMaxOut)
	persistentPeers := conf.GetString(P2pTxPersistentPeers)
	return &p2pConf.P2PConfig{
		AddrBookFilePath: addrFile,
		ListenAddress:    listenAddr,
		MaxConnOutBound:  maxOut,
		MaxConnInBound:   maxIn,
		PersistentPeers:  persistentPeers,
	}
}

func getNodeType(conf *viper.Viper) common.NodeType {
	nodeType := common.NodeType(conf.GetInt(NodeType))
	if (nodeType) <= common.UnknownNode || nodeType >= common.MaxNodeType {
		panic(fmt.Errorf("unknown node type of %d", nodeType))
	}
	return nodeType
}
