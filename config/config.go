package config

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/monitor"
	consensusConfig "github.com/DSiSc/galaxy/consensus/config"
	participatesConfig "github.com/DSiSc/galaxy/participates/config"
	roleConfig "github.com/DSiSc/galaxy/role/config"
	swConf "github.com/DSiSc/gossipswitch/config"
	"github.com/DSiSc/justitia/common"
	"github.com/DSiSc/justitia/tools"
	p2pConf "github.com/DSiSc/p2p/config"
	producerConfig "github.com/DSiSc/producer/config"
	repositoryConfig "github.com/DSiSc/repository/config"
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
	TxpoolSlots    = "general.txpool.globalSlots"
	MaxTxBlock     = "general.txpool.txsPerBlock"
	TxMaxCacheTime = "general.txpool.txMaxCacheTime"
	// consensus policy setting
	ConsensusPolicy                   = "general.consensus.policy"
	ConsensusEnableEmptyBlock         = "general.consensus.enableEmptyBlock"
	ConsensusTimeoutToCollectResponse = "general.consensus.timeoutToCollectResponse"
	ConsensusTimeoutWaitCommit        = "general.consensus.timeoutToWaitCommit"
	ConsensusTimeoutViewChange        = "general.consensus.timeoutToViewChange"
	ConsensusLocalSignatureVerify     = "general.consensus.localSignatureVerify"
	ConsensusSyncSignatureVerify      = "general.consensus.syncSignatureVerify"

	ParticipatesPolicy = "general.participates.policy"
	RolePolicy         = "general.role.policy"
	// node info
	NodeAddress = "general.node.address"
	NodeId      = "general.node.id"
	NodeUrl     = "general.node.url"
	// block chain
	RepositoryPlugin    = "general.repository.plugin"
	RepositoryStatePath = "general.repository.statePath"
	RepositoryDataPath  = "general.repository.dataPath"
	// api gateway
	ApiGatewayAddr = "general.apigateway"
	// Default parameter for solo block producer
	BlockProducedTimeInterval = "general.BlockProducedInterval"

	//P2P Setting
	BlockSyncerP2P     = "general.p2p.blockSyncer" // block syncer p2p config
	BlockP2P           = "general.p2p.block"       // block p2p config
	TxP2P              = "general.p2p.tx"          // tx p2p config
	P2PAddrBook        = "AddrBookFilePath"
	P2PListenAddr      = "ListenAddress"
	P2PMaxOut          = "MaxConnOutBound"
	P2PMaxIn           = "MaxConnInBound"
	P2PPersistendPeers = "PersistentPeers"
	P2PDebug           = "DebugP2P"
	P2PDebugServer     = "DebugServer"
	P2PDebugAddr       = "DebugAddr"
	P2PNAT             = "Nat"
	P2PDisableDNSSeed  = "DisableDNSSeed"
	P2PDNSSeeds        = "DNSSeeds"
	P2PService         = "Service"

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
	ProducerSignatureVerifySwitch  = "general.signature.producer"
	ValidatorSignatureVerifySwitch = "general.signature.validator"

	// switch config
	TxSwitxh                         = "tx_switch"
	TxSwitchSignatureVerifySwitch    = "general.signature.txswitch"
	BlockSwitch                      = "block_switch"
	BlockSwitchSignatureVerifySwitch = "general.signature.blockswitch"
)

type AlgorithmConfig struct {
	//hash algorithm
	HashAlgorithm string
	//signature algorithm
	SignAlgorithm string
}

type SysConfig struct {
	LogLevel log.Level
	LogPath  string
	LogStyle string
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
	// repositoryConfig
	RepositoryConf repositoryConfig.RepositoryConfig
	// Block Produce Interval
	BlockInterval int64
	//algorithm config
	AlgorithmConf AlgorithmConfig
	// producer config
	ProducerConf producerConfig.ProducerConfig

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
	//Switch config
	SwitchConf map[string]*swConf.SwitchConfig
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
	RepositoryConf := NewRepositoryConf(config)
	blockIntervalTime := GetBlockProducerInterval(config)
	prometheusConf := GetPrometheusConf(config)
	expvarConf := GetExpvarConf(config)
	pprofConf := GetPprofConf(config)
	logConf := GetLogSetting(config)
	p2pConf := GetP2PConf(config)
	producerConf := GetProducerConf(config)
	switchConf := GetSwitchConf(config)

	return NodeConfig{
		Account:          nodeAccount,
		NodeType:         nodeType,
		ApiGatewayAddr:   apiGatewayTcpAddr,
		TxPoolConf:       txPoolConf,
		ParticipatesConf: participatesConf,
		RoleConf:         roleConf,
		ConsensusConf:    consensusConf,
		RepositoryConf:   RepositoryConf,
		BlockInterval:    blockIntervalTime,
		AlgorithmConf:    algorithmConf,
		PrometheusConf:   prometheusConf,
		ExpvarConf:       expvarConf,
		PprofConf:        pprofConf,
		Logger:           logConf,
		P2PConf:          p2pConf,
		ProducerConf:     producerConf,
		SwitchConf:       switchConf,
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
	txMaxCacheTime := conf.GetInt64(TxMaxCacheTime)
	txPoolConf := txpool.TxPoolConfig{
		GlobalSlots:    uint64(slots),
		MaxTrsPerBlock: uint64(txPerBlock),
		TxMaxCacheTime: uint64(txMaxCacheTime),
	}
	return txPoolConf
}

func NewParticipateConf(conf *viper.Viper) participatesConfig.ParticipateConfig {
	policy := conf.GetString(ParticipatesPolicy)
	participatesConf := participatesConfig.ParticipateConfig{
		PolicyName: policy,
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
	enableLocalSignatureVerify := conf.GetBool(ConsensusLocalSignatureVerify)
	enableSyncSignatureVerify := conf.GetBool(ConsensusSyncSignatureVerify)
	return consensusConfig.ConsensusConfig{
		PolicyName:       policy,
		EnableEmptyBlock: enableEmptyBlock,
		SignVerifySwitch: consensusConfig.SignatureVerifySwitch{
			LocalVerifySignature: enableLocalSignatureVerify,
			SyncVerifySignature:  enableSyncSignatureVerify,
		},
		Timeout: consensusConfig.ConsensusTimeout{
			TimeoutToCollectResponseMsg: responseTimeout,
			TimeoutToWaitCommitMsg:      commitTimeout,
			TimeoutToChangeView:         viewChangeTimeout,
		},
	}
}

func NewRepositoryConf(conf *viper.Viper) repositoryConfig.RepositoryConfig {
	policy := conf.GetString(RepositoryPlugin)
	dataPath := conf.GetString(RepositoryDataPath)
	statePath := conf.GetString(RepositoryStatePath)
	RepositoryConf := repositoryConfig.RepositoryConfig{
		PluginName:    policy,
		StateDataPath: statePath,
		BlockDataPath: dataPath,
	}
	return RepositoryConf
}

func GetApiGatewayTcpAddr(conf *viper.Viper) string {
	apiGatewayAddr := conf.GetString(ApiGatewayAddr)
	return apiGatewayAddr
}

func GetNodeAccount(conf *viper.Viper) account.Account {
	nodeAddr := conf.GetString(NodeAddress)
	address := tools.HexToAddress(nodeAddr)
	return account.Account{Address: address}
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
	p2pConfig[BlockSyncerP2P] = getP2PConf(BlockSyncerP2P, conf)
	p2pConfig[BlockP2P] = getP2PConf(BlockP2P, conf)
	p2pConfig[TxP2P] = getP2PConf(TxP2P, conf)
	return p2pConfig
}

func getP2PConf(p2pType string, conf *viper.Viper) *p2pConf.P2PConfig {
	addrFile := conf.GetString(p2pType + "." + P2PAddrBook)
	listenAddr := conf.GetString(p2pType + "." + P2PListenAddr)
	maxOut := conf.GetInt(p2pType + "." + P2PMaxOut)
	maxIn := conf.GetInt(p2pType + "." + P2PMaxIn)
	persistentPeers := conf.GetString(p2pType + "." + P2PPersistendPeers)
	debugP2P := conf.GetBool(p2pType + "." + P2PDebug)
	debugServer := conf.GetString(p2pType + "." + P2PDebugServer)
	debugAddr := conf.GetString(p2pType + "." + P2PDebugAddr)
	nat := conf.GetString(p2pType + "." + P2PNAT)
	disableDNSSeed := conf.GetBool(p2pType + "." + P2PDisableDNSSeed)
	dnsSeeds := conf.GetString(p2pType + "." + P2PDNSSeeds)
	service := conf.GetInt(p2pType + "." + P2PService)
	return &p2pConf.P2PConfig{
		AddrBookFilePath: addrFile,
		ListenAddress:    listenAddr,
		MaxConnOutBound:  maxOut,
		MaxConnInBound:   maxIn,
		PersistentPeers:  persistentPeers,
		DebugP2P:         debugP2P,
		DebugServer:      debugServer,
		DebugAddr:        debugAddr,
		NAT:              nat,
		DisableDNSSeed:   disableDNSSeed,
		DNSSeeds:         dnsSeeds,
		Service:          p2pConf.ServiceFlag(service),
	}
}

func GetSwitchConf(conf *viper.Viper) map[string]*swConf.SwitchConfig {
	swConfig := make(map[string]*swConf.SwitchConfig)
	swConfig[TxSwitxh] = getTxSwitchConf(conf)
	swConfig[BlockSwitch] = getBlockSwitchConf(conf)
	return swConfig
}

func getTxSwitchConf(conf *viper.Viper) *swConf.SwitchConfig {
	enableSignVerify := conf.GetBool(TxSwitchSignatureVerifySwitch)
	chainId, _ := GetChainIdFromConfig()
	return &swConf.SwitchConfig{
		VerifySignature: enableSignVerify,
		ChainID:         chainId,
	}
}

func getBlockSwitchConf(conf *viper.Viper) *swConf.SwitchConfig {
	enableSignVerify := conf.GetBool(BlockSwitchSignatureVerifySwitch)
	chainId, _ := GetChainIdFromConfig()
	return &swConf.SwitchConfig{
		VerifySignature: enableSignVerify,
		ChainID:         chainId,
	}
}

func getNodeType(conf *viper.Viper) common.NodeType {
	nodeType := common.NodeType(conf.GetInt(NodeType))
	if (nodeType) <= common.UnknownNode || nodeType >= common.MaxNodeType {
		panic(fmt.Errorf("unknown node type of %d", nodeType))
	}
	return nodeType
}

func GetProducerConf(conf *viper.Viper) producerConfig.ProducerConfig {
	enableSignVerify := conf.GetBool(ProducerSignatureVerifySwitch)
	chainId, _ := GetChainIdFromConfig()
	return producerConfig.ProducerConfig{
		EnableSignatureVerify: enableSignVerify,
		ChainId:               chainId,
	}
}
