package node

import (
	"fmt"
	"github.com/DSiSc/apigateway"
	rpc "github.com/DSiSc/apigateway/rpc/core"
	"github.com/DSiSc/blockchain"
	craftConfig "github.com/DSiSc/craft/config"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/monitor"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy"
	galaxyCommon "github.com/DSiSc/galaxy/common"
	"github.com/DSiSc/galaxy/consensus"
	consensusCommon "github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/policy/dbft"
	"github.com/DSiSc/galaxy/consensus/policy/fbft"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role"
	"github.com/DSiSc/gossipswitch"
	"github.com/DSiSc/gossipswitch/port"
	"github.com/DSiSc/justitia/common"
	"github.com/DSiSc/justitia/config"
	"github.com/DSiSc/justitia/propagator"
	"github.com/DSiSc/justitia/tools"
	"github.com/DSiSc/justitia/tools/events"
	"github.com/DSiSc/p2p"
	"github.com/DSiSc/producer"
	"github.com/DSiSc/syncer"
	"github.com/DSiSc/txpool"
	"github.com/DSiSc/validator"
	"github.com/DSiSc/validator/tools/account"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type NodesService interface {
	Start()
	Stop() error
	Wait()
	Restart() error
}

// node struct with all service
type Node struct {
	nodeWg          sync.WaitGroup
	config          config.NodeConfig
	txpool          txpool.TxsPool
	participates    participates.Participates
	role            role.Role
	consensus       consensus.Consensus
	producer        *producer.Producer
	txSwitch        *gossipswitch.GossipSwitch
	blockSwitch     *gossipswitch.GossipSwitch
	validator       *validator.Validator
	rpcListeners    []net.Listener
	eventCenter     types.EventCenter
	msgChannel      chan common.MsgType
	serviceChannel  chan interface{}
	blockSyncerP2P  p2p.P2PAPI
	blockSyncer     syncer.BlockSyncerAPI
	blockP2P        p2p.P2PAPI
	blockPropagator *propagator.BlockPropagator
	txP2P           p2p.P2PAPI
	txPropagator    *propagator.TxPropagator
}

func InitLog(args common.SysConfig, conf config.NodeConfig) {
	var logPath = args.LogPath
	if common.BlankString != logPath {
		conf.Logger.Appenders[config.FileLogAppender].Enabled = true
		conf.Logger.Appenders[config.FileLogAppender].LogPath = logPath
	}
	var logFormat = args.LogStyle
	if common.BlankString != logFormat {
		conf.Logger.Appenders[config.FileLogAppender].Enabled = true
		conf.Logger.Appenders[config.FileLogAppender].Format = logFormat
	}
	var logLevel = args.LogLevel
	if common.InvalidInt != int(logLevel) {
		conf.Logger.Appenders[config.FileLogAppender].Enabled = true
		conf.Logger.Appenders[config.FileLogAppender].LogLevel = log.Level(uint8(logLevel))
	}

	if conf.Logger.Appenders[config.FileLogAppender].Enabled {
		// initialize logfile
		logPath = conf.Logger.Appenders[config.FileLogAppender].LogPath
		tools.EnsureFolderExist(logPath[0:strings.LastIndex(logPath, "/")])
		logfile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
		conf.Logger.Appenders[config.FileLogAppender].Output = logfile
	}

	log.SetGlobalConfig(&conf.Logger)
}

func NewNode(args common.SysConfig) (NodesService, error) {
	nodeConf := config.NewNodeConfig()
	InitLog(args, nodeConf)
	craftConfig.GlobalConfig.Store(craftConfig.HashAlgName, nodeConf.AlgorithmConf.HashAlgorithm)
	pool := txpool.NewTxPool(nodeConf.TxPoolConf)
	eventsCenter := events.NewEvent()
	txSwitch, err := gossipswitch.NewGossipSwitchByType(gossipswitch.TxSwitch, eventsCenter, nodeConf.SwitchConf[config.TxSwitxh])
	if err != nil {
		log.Error("Init txSwitch failed.")
		return nil, fmt.Errorf("txswitch init failed")
	}
	swChIn := txSwitch.InPort(port.LocalInPortId).Channel()
	rpc.SetSwCh(swChIn)
	err = txSwitch.OutPort(port.LocalInPortId).BindToPort(func(msg interface{}) error {
		return pool.AddTx(msg.(*types.Transaction))
	})
	if err != nil {
		log.Error("Register txpool failed.")
		return nil, fmt.Errorf("registe txpool failed")
	}
	blkSwitch, err := gossipswitch.NewGossipSwitchByType(gossipswitch.BlockSwitch, eventsCenter, nodeConf.SwitchConf[config.BlockSwitch])
	if err != nil {
		log.Error("Init block switch failed.")
		return nil, fmt.Errorf("blkSwitch init failed")
	}
	err = blockchain.InitBlockChain(nodeConf.BlockChainConf, eventsCenter)
	if err != nil {
		log.Error("Init block chain failed.")
		return nil, fmt.Errorf("blockchain init failed")
	}
	blockSyncerP2P, err := p2p.NewP2P(nodeConf.P2PConf[config.BlockSyncerP2P], eventsCenter)
	if err != nil {
		log.Error("Init block syncer p2p failed.")
		return nil, fmt.Errorf("init block syncer p2p failed")
	}
	blockSyncer, err := syncer.NewBlockSyncer(blockSyncerP2P, blkSwitch.InPort(port.LocalInPortId).Channel(), eventsCenter)
	if err != nil {
		log.Error("Init block syncer failed.")
		return nil, fmt.Errorf("init block syncer failed")
	}
	blockP2P, err := p2p.NewP2P(nodeConf.P2PConf[config.BlockP2P], eventsCenter)
	if err != nil {
		log.Error("Init block p2p failed.")
		return nil, fmt.Errorf("init block p2p failed")
	}
	blockPropagator, err := propagator.NewBlockPropagator(blockP2P, blkSwitch.InPort(port.RemoteInPortId).Channel(), eventsCenter)
	if err != nil {
		log.Error("Init block propagator failed.")
		return nil, fmt.Errorf("init block propagator failed")
	}
	txP2P, err := p2p.NewP2P(nodeConf.P2PConf[config.TxP2P], eventsCenter)
	if err != nil {
		log.Error("Init tx p2p failed.")
		return nil, fmt.Errorf("init tx p2p failed")
	}
	txPropagator, err := propagator.NewTxPropagator(txP2P, txSwitch.InPort(port.RemoteInPortId).Channel())
	if err != nil {
		log.Error("Init tx propagator failed.")
		return nil, fmt.Errorf("init tx propagator failed")
	}
	txSwitch.OutPort(port.RemoteOutPortId).BindToPort(txPropagator.TxSwitchOutPutFunc())
	node := &Node{
		config:          nodeConf,
		txpool:          pool,
		txSwitch:        txSwitch,
		blockSwitch:     blkSwitch,
		eventCenter:     eventsCenter,
		msgChannel:      make(chan common.MsgType),
		serviceChannel:  make(chan interface{}),
		blockSyncerP2P:  blockSyncerP2P,
		blockSyncer:     blockSyncer,
		blockP2P:        blockP2P,
		blockPropagator: blockPropagator,
		txP2P:           txP2P,
		txPropagator:    txPropagator,
	}
	if common.ConsensusNode == nodeConf.NodeType {
		galaxyConfig := galaxyCommon.GalaxyPluginConf{
			Account:         nodeConf.Account,
			BlockSwitch:     blkSwitch.InPort(port.LocalInPortId).Channel(),
			ParticipateConf: nodeConf.ParticipatesConf,
			RoleConf:        nodeConf.RoleConf,
			ConsensusConf:   nodeConf.ConsensusConf,
		}
		galaxyPlugin, err := galaxy.NewGalaxyPlugin(galaxyConfig)
		if err != nil {
			log.Error("Init galaxy plugin failed.")
			return node, fmt.Errorf("init galaxy plugin failed with error %v", err)
		}
		node.participates = galaxyPlugin.Participates
		node.role = galaxyPlugin.Role
		node.consensus = galaxyPlugin.Consensus
	}
	node.eventsRegister()
	return node, nil
}

func (instance *Node) eventsRegister() {
	instance.eventCenter.Subscribe(types.EventBlockCommitted, func(v interface{}) {
		if nil != v {
			block := v.(*types.Block)
			log.Debug("begin delete txs after block %d committed success.", block.Header.Height)
			instance.txpool.DelTxs(block.Transactions)
		}
	})
	if common.ConsensusNode == instance.config.NodeType {
		instance.eventCenter.Subscribe(types.EventBlockCommitted, func(v interface{}) {
			instance.msgChannel <- common.MsgBlockCommitSuccess
		})
		instance.eventCenter.Subscribe(types.EventBlockVerifyFailed, func(v interface{}) {
			instance.msgChannel <- common.MsgBlockVerifyFailed
		})
		instance.eventCenter.Subscribe(types.EventBlockCommitFailed, func(v interface{}) {
			instance.msgChannel <- common.MsgBlockCommitFailed
		})
		instance.eventCenter.Subscribe(types.EventConsensusFailed, func(v interface{}) {
			instance.msgChannel <- common.MsgToConsensusFailed
		})
		instance.eventCenter.Subscribe(types.EventMasterChange, func(v interface{}) {
			instance.msgChannel <- common.MsgChangeMaster
		})
		instance.eventCenter.Subscribe(types.EventOnline, func(v interface{}) {
			instance.msgChannel <- common.MsgOnline
		})
		instance.eventCenter.Subscribe(types.EventBlockWithoutTxs, func(v interface{}) {
			instance.msgChannel <- common.MsgBlockWithoutTx
		})
	}
}

func (instance *Node) eventUnregister() {
	instance.eventCenter.UnSubscribeAll()
}

func (instance *Node) notify() {
	go func() {
		instance.msgChannel <- common.MsgRoundRunFailed
	}()
}

func (instance *Node) blockFactory(master account.Account, participates []account.Account) {
	monitor.JTMetrics.ConsensusPeerId.Set(float64(instance.config.Account.Extension.Id))
	monitor.JTMetrics.ConsensusMasterId.Set(float64(master.Extension.Id))
	instance.consensus.Initialization(master, participates, instance.eventCenter, false)
	isMaster := master == instance.config.Account
	if isMaster {
		log.Info("Master this round.")
		if nil == instance.producer {
			instance.producer = producer.NewProducer(instance.txpool, instance.config.Account, instance.config.ProducerConf)
		}
		block, err := instance.producer.MakeBlock()
		if err != nil {
			log.Error("Make block failed with err %v.", err)
			instance.notify()
			return
		}
		proposal := &consensusCommon.Proposal{
			Block: block,
		}
		if err = instance.consensus.ToConsensus(proposal); err != nil {
			log.Error("ToConsensus failed with err %v.", err)
		} else {
			log.Info("Block has been confirmed with height %d and hash %x.",
				block.Header.Height, block.HeaderHash)
		}
	} else {
		log.Info("Slave this round.")
	}
}

func (instance *Node) NextRound(msgType common.MsgType) {
	switch instance.consensus.(type) {
	case *dbft.DBFTPolicy:
		if common.MsgChangeMaster == msgType {
			consensusResult := instance.consensus.GetConsensusResult()
			instance.blockFactory(consensusResult.Master, consensusResult.Participate)
		} else {
			instance.Round()
		}
	case *fbft.FBFTPolicy:
		consensusResult := instance.consensus.GetConsensusResult()
		log.Debug("get participate %v and master %v.",
			consensusResult.Participate, consensusResult.Master.Extension.Id)
		if common.MsgBlockCommitSuccess == msgType {
			time.Sleep(time.Duration(instance.config.BlockInterval) * time.Millisecond)
		}
		instance.blockFactory(consensusResult.Master, consensusResult.Participate)
	default:
		instance.Round()
	}
}

func (instance *Node) Round() {
	log.Debug("start a new round.")
	time.Sleep(time.Duration(instance.config.BlockInterval) * time.Millisecond)
	participate, err := instance.participates.GetParticipates()
	if err != nil {
		log.Error("get participates failed with error %s.", err)
		instance.notify()
		return
	}
	_, master, err := instance.role.RoleAssignments(participate)
	if nil != err {
		log.Error("Role assignments failed with err %v.", err)
		instance.notify()
		return
	}
	instance.blockFactory(master, participate)
}

func (instance *Node) OnlineWizard() {
	log.Info("start online wizard.")
	participate, err := instance.participates.GetParticipates()
	if err != nil {
		log.Error("get participates failed with error %s.", err)
		instance.notify()
		return
	}
	_, master, err := instance.role.RoleAssignments(participate)
	if nil != err {
		log.Error("Role assignments failed with err %v.", err)
		instance.notify()
		return
	}
	instance.consensus.Initialization(master, participate, instance.eventCenter, false)
	instance.consensus.Online()
}

func (instance *Node) mainLoop() {
	instance.OnlineWizard()
	for {
		msg := <-instance.msgChannel
		switch msg {
		case common.MsgBlockCommitSuccess:
			log.Info("Receive msg from switch is success.")
			instance.NextRound(common.MsgBlockCommitSuccess)
		case common.MsgBlockCommitFailed:
			instance.NextRound(common.MsgBlockCommitFailed)
			log.Info("Receive msg from switch is commit failed.")
		case common.MsgBlockVerifyFailed:
			log.Info("Receive msg from switch is verify failed.")
			instance.NextRound(common.MsgBlockVerifyFailed)
		case common.MsgRoundRunFailed:
			log.Error("Receive msg from main loop is run failed.")
			instance.NextRound(common.MsgBlockVerifyFailed)
		case common.MsgToConsensusFailed:
			log.Error("Receive msg of to consensus failed.")
			instance.NextRound(common.MsgToConsensusFailed)
		case common.MsgChangeMaster:
			log.Info("Receive msg of change views.")
			instance.NextRound(common.MsgBlockCommitSuccess)
		case common.MsgOnline:
			log.Info("Receive msg of online.")
			instance.NextRound(common.MsgOnline)
		case common.MsgBlockWithoutTx:
			log.Info("Receive msg of block without transaction.")
			instance.NextRound(common.MsgBlockCommitSuccess)
		case common.MsgNodeServiceStopped:
			log.Warn("Stop node service.")
			break
		}
	}
}

func (instance *Node) stratRpc() {
	var err error
	if instance.rpcListeners, err = apigateway.StartRPC(instance.config.ApiGatewayAddr); nil != err {
		panic(fmt.Sprintf("Rpc start failed with %v.", err))
	}
}

func (instance *Node) startSwitch() {
	if err := instance.txSwitch.Start(); nil != err {
		panic(fmt.Sprintf("TxSwitch start failed with %v.", err))
	}
	if err := instance.blockSwitch.Start(); nil != err {
		panic(fmt.Sprintf("BlockSwitch start failed with %v.", err))
	}
}

func (instance *Node) startBlockSyncer() {
	if err := instance.blockSyncerP2P.Start(); nil != err {
		panic(fmt.Sprintf("Start block syncer p2p failed with error %v.", err))
	}
	if err := instance.blockSyncer.Start(); nil != err {
		panic(fmt.Sprintf("Start block syncer failed with error %v.", err))
	}
}

func (instance *Node) startBlockPropagator() {
	if err := instance.blockP2P.Start(); nil != err {
		panic(fmt.Sprintf("Start block p2p failed with error %v.", err))
	}
	if err := instance.blockPropagator.Start(); nil != err {
		panic(fmt.Sprintf("Start block propagator failed with error %v.", err))
	}
}

func (instance *Node) startTxPropagator() {
	if err := instance.txP2P.Start(); nil != err {
		panic(fmt.Sprintf("Start tx p2p failed with error %v.", err))
	}
	if err := instance.txPropagator.Start(); nil != err {
		panic(fmt.Sprintf("Start tx propagator failed with error %v.", err))
	}
}

func (instance *Node) Start() {
	instance.stratRpc()
	instance.startSwitch()
	instance.startBlockSyncer()
	instance.startBlockPropagator()
	instance.startTxPropagator()
	monitor.StartPrometheusServer(instance.config.PrometheusConf)
	monitor.StartExpvarServer(instance.config.ExpvarConf)
	monitor.StartPprofServer(instance.config.PprofConf)
	if instance.config.NodeType == common.ConsensusNode {
		go instance.consensus.Start()
		go instance.mainLoop()
	}
}

func (instance *Node) Stop() error {
	log.Warn("Stop node service.")
	close(instance.serviceChannel)
	var err error
	for _, listener := range instance.rpcListeners {
		if err := listener.Close(); err != nil {
			log.Error("Stop rpc listeners failed with error %v.", err)
			continue
		}
	}
	instance.blockSyncerP2P.Stop()
	instance.blockSyncer.Stop()
	instance.blockP2P.Stop()
	instance.blockPropagator.Stop()
	instance.txP2P.Stop()
	instance.txPropagator.Stop()
	instance.blockSwitch.Stop()
	instance.txSwitch.Stop()
	instance.eventUnregister()
	if instance.config.NodeType == common.ConsensusNode {
		instance.msgChannel <- common.MsgNodeServiceStopped
		monitor.StopPrometheusServer()
	}
	return err
}

func (instance *Node) Wait() {
	<-instance.serviceChannel
}

func (instance *Node) Restart() error {
	if err := instance.Stop(); err != nil {
		log.Error("restart service failed with err %v.", err)
		return err
	}
	instance.Start()
	return nil
}
