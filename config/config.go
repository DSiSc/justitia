package config

import (
	"fmt"
	blockchainConfig "github.com/DSiSc/blockchain/config"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/monitor"
	consensusConfig "github.com/DSiSc/galaxy/consensus/config"
	participatesCommon "github.com/DSiSc/galaxy/participates/common"
	participatesConfig "github.com/DSiSc/galaxy/participates/config"
	roleConfig "github.com/DSiSc/galaxy/role/config"
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
	TxpoolSlots = "general.txpool.globalSlots"
	MaxTxBlock  = "general.txpool.txsPerBlock"
	// consensus policy setting
	ConsensusPolicy                   = "general.consensus.policy"
	ConsensusEnableEmptyBlock         = "general.consensus.enableEmptyBlock"
	ConsensusTimeoutToCollectResponse = "general.consensus.timeoutToCollectResponse"
	ConsensusTimeoutWaitCommit        = "general.consensus.timeoutToWaitCommit"
	ConsensusTimeoutViewChange        = "general.consensus.timeoutToViewChange"

	ParticipatesPolicy   = "general.participates.policy"
	ParticipatesNumber   = "general.participates.participates"
	ParticipatesNodeInfo = "general.participates.node"
	RolePolicy           = "general.role.policy"
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
	BlockProducedTimeInterval = "general.BlockProducedInterval"

	//P2P Setting
	// block syncer p2p config
	BlockSyncerP2P                = "block_syncer_p2p"
	BlockSyncerP2PAddrBook        = "general.p2p.blockSyncer.AddrBookFilePath"
	BlockSyncerP2PListenAddr      = "general.p2p.blockSyncer.ListenAddress"
	BlockSyncerP2PMaxOut          = "general.p2p.blockSyncer.MaxConnOutBound"
	BlockSyncerP2PMaxIn           = "general.p2p.blockSyncer.MaxConnInBound"
	BlockSyncerP2PPersistendPeers = "general.p2p.blockSyncer.PersistentPeers"
	BlockSyncerP2PDebug           = "general.p2p.blockSyncer.DebugP2P"
	BlockSyncerP2PDebugServer     = "general.p2p.blockSyncer.DebugServer"
	BlockSyncerP2PDebugAddr       = "general.p2p.blockSyncer.DebugAddr"

	// block p2p config
	BlockP2P                = "block_p2p"
	BlockP2PAddrBook        = "general.p2p.block.AddrBookFilePath"
	BlockP2PListenAddr      = "general.p2p.block.ListenAddress"
	BlockP2PMaxOut          = "general.p2p.block.MaxConnOutBound"
	BlockP2PMaxIn           = "general.p2p.block.MaxConnInBound"
	BlockP2PPersistendPeers = "general.p2p.block.PersistentPeers"
	BlockP2PDebug           = "general.p2p.block.DebugP2P"
	BlockP2PDebugServer     = "general.p2p.block.DebugServer"
	BlockP2PDebugAddr       = "general.p2p.block.DebugAddr"

	// tx p2p config
	TxP2P                = "tx_p2p"
	TxP2PAddrBook        = "general.p2p.tx.AddrBookFilePath"
	TxP2PListenAddr      = "general.p2p.tx.ListenAddress"
	TxP2PMaxOut          = "general.p2p.tx.MaxConnOutBound"
	TxP2PMaxIn           = "general.p2p.tx.MaxConnInBound"
	TxP2PPersistendPeers = "general.p2p.tx.PersistentPeers"
	TxP2PDebug           = "general.p2p.tx.DebugP2P"
	TxP2PDebugServer     = "general.p2p.tx.DebugServer"
	TxP2PDebugAddr       = "general.p2p.tx.DebugAddr"

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
	ConsoleLogAppender = "logging.console"
	LogConsoleEnabled  = "logging.console.enabled"
	LogConsoleLevel    = "logging.console.level"
	LogConsoleFormat   = "logging.console.format"
	LogConsoleCaller   = "logging.console.caller"
	LogConsoleHostname = "logging.console.hostname"
	FileLogAppender    = "logging.file"
	LogFileEnabled     = "logging.file.enabled"
	LogFilePath        = "logging.file.path"
	LogFileLevel       = "logging.file.level"
	LogFileFormat      = "logging.file.format"
	LogFileCaller      = "logging.file.caller"
	LogFileHostname    = "logging.file.hostname"

	// signature switch
	ProducerSignatureVerifySwitch    = "general.signature.producer"
	TxSwitchSignatureVerifySwitch    = "general.signature.txswitch"
	ValidatorSignatureVerifySwitch   = "general.signature.validator"
	BlockSwitchSignatureVerifySwitch = "general.signature.blockswitch"
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
	ParticipatesConf participatesConfig.ParticipateConfig
	// role
	RoleConf roleConfig.RoleConfig
	// consensus
	ConsensusConf consensusConfig.ConsensusConfig
	// BlockChainConfig
	BlockChainConf blockchainConfig.BlockChainConfig
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
	p2pConf := GetP2PConf(config)

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
		P2PConf:          p2pConf,
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

func NewParticipateConf(conf *viper.Viper) participatesConfig.ParticipateConfig {
	policy := conf.GetString(ParticipatesPolicy)
	participates := conf.GetInt64(ParticipatesNumber)
	participatesConf := participatesConfig.ParticipateConfig{
		PolicyName: policy,
		Delegates:  uint64(participates),
	}
	if policy != participatesCommon.SoloPolicy {
		accounts := make([]account.Account, 0)
		for index := int64(0); index < participates; index++ {
			nodePath := fmt.Sprintf("%s%d", ParticipatesNodeInfo, index)
			addressPath := nodePath + ".address"
			address := conf.GetString(addressPath)
			idPath := nodePath + ".id"
			id := conf.GetInt64(idPath)
			urlPath := nodePath + ".url"
			url := conf.GetString(urlPath)
			accounts = append(accounts, account.Account{
				Address: tools.HexToAddress(address),
				Extension: account.AccountExtension{
					Id:  uint64(id),
					Url: url,
				},
			})
		}
		participatesConf.Participates = accounts
	}
	return participatesConf
}

func NewRoleConf(conf *viper.Viper) roleConfig.RoleConfig {
	policy := conf.GetString(RolePolicy)
	roleConf := roleConfig.RoleConfig{
		PolicyName: policy,
	}
	return roleConf
}

func NewConsensusConf(conf *viper.Viper) consensusConfig.ConsensusConfig {
	policy := conf.GetString(ConsensusPolicy)
	responseTimeout := conf.GetInt64(ConsensusTimeoutToCollectResponse)
	commitTimeout := conf.GetInt64(ConsensusTimeoutWaitCommit)
	viewChangeTimeout := conf.GetInt64(ConsensusTimeoutViewChange)
	enableEmptyBlock := conf.GetBool(ConsensusEnableEmptyBlock)
	return consensusConfig.ConsensusConfig{
		PolicyName:       policy,
		EnableEmptyBlock: enableEmptyBlock,
		Timeout: consensusConfig.ConsensusTimeout{
			TimeoutToCollectResponseMsg: responseTimeout,
			TimeoutToWaitCommitMsg:      commitTimeout,
			TimeoutToChangeView:         viewChangeTimeout,
		},
	}
}

func NewBlockChainConf(conf *viper.Viper) blockchainConfig.BlockChainConfig {
	policy := conf.GetString(BlockChainPlugin)
	dataPath := conf.GetString(BlockChainDataPath)
	statePath := conf.GetString(BlockChainStatePath)
	blockChainConf := blockchainConfig.BlockChainConfig{
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
	blockInterval := conf.GetInt64(BlockProducedTimeInterval)
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
	logConsoleFormat := conf.GetString(LogConsoleFormat)
	logConsoleCaller := conf.GetBool(LogConsoleCaller)
	logConsoleHostname := conf.GetBool(LogConsoleHostname)
	logFileEnabled := conf.GetBool(LogFileEnabled)
	logFilePath := conf.GetString(LogFilePath)
	logFileLevel := conf.GetInt(LogFileLevel)
	logFileFormat := conf.GetString(LogFileFormat)
	logFileCaller := conf.GetBool(LogFileCaller)
	logFileHostname := conf.GetBool(LogFileHostname)

	consoleAppender := &log.Appender{
		Enabled:      logConsoleEnabled,
		LogLevel:     log.Level(logConsoleLevel),
		LogType:      log.ConsoleLog,
		LogPath:      log.ConsoleStdout,
		Output:       os.Stdout,
		Format:       strings.ToUpper(logConsoleFormat),
		ShowCaller:   logConsoleCaller,
		ShowHostname: logConsoleHostname,
	}
	//tools.EnsureFolderExist(logFilePath[0:strings.LastIndex(logFilePath, "/")])
	//logfile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	//if err != nil {
	//	panic(err)
	//}
	fileAppender := &log.Appender{
		Enabled:      logFileEnabled,
		LogLevel:     log.Level(logFileLevel),
		LogType:      log.FileLog,
		LogPath:      logFilePath,
		Output:       nil,
		Format:       strings.ToUpper(logFileFormat),
		ShowCaller:   logFileCaller,
		ShowHostname: logFileHostname,
	}

	globalLogConfig := log.Config{
		Enabled:         logConsoleEnabled || logFileEnabled,
		Provider:        log.GetGlobalConfig().Provider,
		GlobalLogLevel:  log.Level(uint8(math.Max(float64(logConsoleLevel), float64(logFileLevel)))),
		TimeFieldFormat: logTimestampFormat,
		Appenders:       map[string]*log.Appender{ConsoleLogAppender: consoleAppender, FileLogAppender: fileAppender},
		OutputFlags:     log.GetOutputFlags(),
	}
	return globalLogConfig
}

func GetP2PConf(conf *viper.Viper) map[string]*p2pConf.P2PConfig {
	p2pConfig := make(map[string]*p2pConf.P2PConfig)
	p2pConfig[BlockSyncerP2P] = getBlockSyncerP2PConf(conf)
	p2pConfig[BlockP2P] = getBlockP2PConf(conf)
	p2pConfig[TxP2P] = getTxP2PConf(conf)
	return p2pConfig
}

func getBlockSyncerP2PConf(conf *viper.Viper) *p2pConf.P2PConfig {
	addrFile := conf.GetString(BlockSyncerP2PAddrBook)
	listenAddr := conf.GetString(BlockSyncerP2PListenAddr)
	maxOut := conf.GetInt(BlockSyncerP2PMaxOut)
	maxIn := conf.GetInt(BlockSyncerP2PMaxIn)
	persistentPeers := conf.GetString(BlockSyncerP2PPersistendPeers)
	debugP2P := conf.GetBool(BlockSyncerP2PDebug)
	debugServer := conf.GetString(BlockSyncerP2PDebugServer)
	debugAddr := conf.GetString(BlockSyncerP2PDebugAddr)
	return &p2pConf.P2PConfig{
		AddrBookFilePath: addrFile,
		ListenAddress:    listenAddr,
		MaxConnOutBound:  maxOut,
		MaxConnInBound:   maxIn,
		PersistentPeers:  persistentPeers,
		DebugP2P:         debugP2P,
		DebugServer:      debugServer,
		DebugAddr:        debugAddr,
	}
}

func getBlockP2PConf(conf *viper.Viper) *p2pConf.P2PConfig {
	addrFile := conf.GetString(BlockP2PAddrBook)
	listenAddr := conf.GetString(BlockP2PListenAddr)
	maxOut := conf.GetInt(BlockP2PMaxOut)
	maxIn := conf.GetInt(BlockP2PMaxIn)
	persistentPeers := conf.GetString(BlockP2PPersistendPeers)
	debugP2P := conf.GetBool(BlockP2PDebug)
	debugServer := conf.GetString(BlockP2PDebugServer)
	debugAddr := conf.GetString(BlockP2PDebugAddr)
	return &p2pConf.P2PConfig{
		AddrBookFilePath: addrFile,
		ListenAddress:    listenAddr,
		MaxConnOutBound:  maxOut,
		MaxConnInBound:   maxIn,
		PersistentPeers:  persistentPeers,
		DebugP2P:         debugP2P,
		DebugServer:      debugServer,
		DebugAddr:        debugAddr,
	}
}

func getTxP2PConf(conf *viper.Viper) *p2pConf.P2PConfig {
	addrFile := conf.GetString(TxP2PAddrBook)
	listenAddr := conf.GetString(TxP2PListenAddr)
	maxOut := conf.GetInt(TxP2PMaxOut)
	maxIn := conf.GetInt(TxP2PMaxIn)
	persistentPeers := conf.GetString(TxP2PPersistendPeers)
	debugP2P := conf.GetBool(TxP2PDebug)
	debugServer := conf.GetString(TxP2PDebugServer)
	debugAddr := conf.GetString(TxP2PDebugAddr)
	return &p2pConf.P2PConfig{
		AddrBookFilePath: addrFile,
		ListenAddress:    listenAddr,
		MaxConnOutBound:  maxOut,
		MaxConnInBound:   maxIn,
		PersistentPeers:  persistentPeers,
		DebugP2P:         debugP2P,
		DebugServer:      debugServer,
		DebugAddr:        debugAddr,
	}
}

func getNodeType(conf *viper.Viper) common.NodeType {
	nodeType := common.NodeType(conf.GetInt(NodeType))
	if (nodeType) <= common.UnknownNode || nodeType >= common.MaxNodeType {
		panic(fmt.Errorf("unknown node type of %d", nodeType))
	}
	return nodeType
}
