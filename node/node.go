package node

import (
	"fmt"
	"github.com/DSiSc/apigateway"
	rpc "github.com/DSiSc/apigateway/rpc/core"
	"github.com/DSiSc/blockchain"
	gconfing "github.com/DSiSc/craft/config"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/monitor"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy"
	galaxyCommon "github.com/DSiSc/galaxy/common"
	"github.com/DSiSc/galaxy/consensus"
	commonc "github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/policy/dbft"
	"github.com/DSiSc/galaxy/consensus/policy/fbft"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role"
	"github.com/DSiSc/gossipswitch"
	"github.com/DSiSc/gossipswitch/port"
	"github.com/DSiSc/justitia/common"
	"github.com/DSiSc/justitia/config"
	"github.com/DSiSc/justitia/propagator"
	"github.com/DSiSc/justitia/tools/events"
	"github.com/DSiSc/p2p"
	"github.com/DSiSc/producer"
	"github.com/DSiSc/syncer"
	"github.com/DSiSc/txpool"
	"github.com/DSiSc/validator"
	"github.com/DSiSc/validator/tools/account"
	"net"
	"os"
	"sync"
	"time"
)

type NodeService interface {
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
		logfile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
		conf.Logger.Appenders["filelog"].Output = logfile
	}
	var logFormat = args.LogStyle
	if common.BlankString != logFormat {
		conf.Logger.Appenders["filelog"].Format = logFormat
	}
	var logLevel = args.LogLevel
	if common.InvalidInt != int(logLevel) {
		conf.Logger.Appenders["filelog"].LogLevel= log.Level(uint8(logLevel))
	}

	log.SetGlobalConfig(&conf.Logger)
}

func NewNode(args common.SysConfig) (NodeService, error) {
	nodeConf := config.NewNodeConfig()
	InitLog(args, nodeConf)
	gconfing.GlobalConfig.Store(gconfing.HashAlgName, nodeConf.AlgorithmConf.HashAlgorithm)
	pool := txpool.NewTxPool(nodeConf.TxPoolConf)
	eventsCenter := events.NewEvent()
	txSwitch, err := gossipswitch.NewGossipSwitchByType(gossipswitch.TxSwitch, eventsCenter)
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
	blkSwitch, err := gossipswitch.NewGossipSwitchByType(gossipswitch.BlockSwitch, eventsCenter)
	if err != nil {
		log.Error("Init block switch failed.")
		return nil, fmt.Errorf("blkSwitch init failed")
	}
	err = blockchain.InitBlockChain(nodeConf.BlockChainConf, eventsCenter)
	if err != nil {
		log.Error("Init block chain failed.")
		return nil, fmt.Errorf("blockchain init failed")
	}
	blockSyncerP2P, err := p2p.NewP2P(nodeConf.P2PConf[config.BLOCK_SYNCER_P2P], eventsCenter)
	if err != nil {
		log.Error("Init block syncer p2p failed.")
		return nil, fmt.Errorf("Init block syncer p2p failed. ")
	}
	blockSyncer, err := syncer.NewBlockSyncer(blockSyncerP2P, blkSwitch.InPort(port.LocalInPortId).Channel(), eventsCenter)
	if err != nil {
		log.Error("Init block syncer failed.")
		return nil, fmt.Errorf("Init block syncer failed. ")
	}
	blockP2P, err := p2p.NewP2P(nodeConf.P2PConf[config.BLOCK_P2P], eventsCenter)
	if err != nil {
		log.Error("Init block p2p failed.")
		return nil, fmt.Errorf("Init block p2p failed. ")
	}
	blockPropagator, err := propagator.NewBlockPropagator(blockP2P, blkSwitch.InPort(port.RemoteInPortId).Channel(), eventsCenter)
	if err != nil {
		log.Error("Init block propagator failed.")
		return nil, fmt.Errorf("Init block propagator failed. ")
	}
	txP2P, err := p2p.NewP2P(nodeConf.P2PConf[config.TX_P2P], eventsCenter)
	if err != nil {
		log.Error("Init tx p2p failed.")
		return nil, fmt.Errorf("Init tx p2p failed. ")
	}
	txPropagator, err := propagator.NewTxPropagator(txP2P, txSwitch.InPort(port.RemoteInPortId).Channel())
	if err != nil {
		log.Error("Init tx propagator failed.")
		return nil, fmt.Errorf("Init tx propagator failed. ")
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

func (self *Node) eventsRegister() {
	self.eventCenter.Subscribe(types.EventBlockCommitted, func(v interface{}) {
		if nil != v {
			block := v.(*types.Block)
			log.Info("begin delete txs after block %d committed success.", block.Header.Height)
			self.txpool.DelTxs(block.Transactions)
		}
	})
	if common.ConsensusNode == self.config.NodeType {
		self.eventCenter.Subscribe(types.EventBlockCommitted, func(v interface{}) {
			self.msgChannel <- common.MsgBlockCommitSuccess
		})
		self.eventCenter.Subscribe(types.EventBlockVerifyFailed, func(v interface{}) {
			self.msgChannel <- common.MsgBlockVerifyFailed
		})
		self.eventCenter.Subscribe(types.EventBlockCommitFailed, func(v interface{}) {
			self.msgChannel <- common.MsgBlockCommitFailed
		})
		self.eventCenter.Subscribe(types.EventConsensusFailed, func(v interface{}) {
			self.msgChannel <- common.MsgToConsensusFailed
		})
		self.eventCenter.Subscribe(types.EventMasterChange, func(v interface{}) {
			self.msgChannel <- common.MsgChangeMaster
		})
		self.eventCenter.Subscribe(types.EventOnline, func(v interface{}) {
			self.msgChannel <- common.MsgOnline
		})
	}
}

func (self *Node) eventUnregister() {
	self.eventCenter.UnSubscribeAll()
}

func (self *Node) notify() {
	go func() {
		self.msgChannel <- common.MsgRoundRunFailed
	}()
}

func (self *Node) blockFactory(master account.Account, participates []account.Account) {
	self.consensus.Initialization(master, participates, self.eventCenter, false)
	isMaster := master == self.config.Account
	if isMaster {
		log.Info("Master this round.")
		if nil == self.producer {
			self.producer = producer.NewProducer(self.txpool, &self.config.Account)
		}
		block, err := self.producer.MakeBlock()
		if err != nil {
			log.Error("Make block failed with err %v.", err)
			self.notify()
			return
		}
		proposal := &commonc.Proposal{
			Block: block,
		}
		if err = self.consensus.ToConsensus(proposal); err != nil {
			log.Error("ToConsensus failed with err %v.", err)
		} else {
			log.Info("Block has been confirmed with height %d and hash %x.",
				block.Header.Height, block.HeaderHash)
		}
	} else {
		log.Info("Slave this round.")
		if self.validator == nil {
			log.Info("validator is nil.")
			self.validator = validator.NewValidator(&self.config.Account)
		}
	}
}

func (self *Node) NextRound(msgType common.MsgType) {
	switch self.consensus.(type) {
	case *dbft.DBFTPolicy:
		if common.MsgChangeMaster == msgType {
			consensusResult := self.consensus.GetConsensusResult()
			self.blockFactory(consensusResult.Master, consensusResult.Participate)
		} else {
			self.Round()
		}
	case *fbft.FBFTPolicy:
		consensusResult := self.consensus.GetConsensusResult()
		log.Debug("get participate %v and master %v.",
			consensusResult.Participate, consensusResult.Master.Extension.Id)
		if common.MsgBlockCommitSuccess == msgType {
			time.Sleep(common.FBFTRoundInterval * time.Millisecond)
		}
		self.blockFactory(consensusResult.Master, consensusResult.Participate)
	default:
		self.Round()
	}
}

func (self *Node) Round() {
	log.Debug("start a new round.")
	time.Sleep(time.Duration(self.config.BlockInterval) * time.Millisecond)
	participates, err := self.participates.GetParticipates()
	if err != nil {
		log.Error("get participates failed with error %s.", err)
		self.notify()
		return
	}
	_, master, err := self.role.RoleAssignments(participates)
	if nil != err {
		log.Error("Role assignments failed with err %v.", err)
		self.notify()
		return
	}
	self.blockFactory(master, participates)
}

func (self *Node) OnlineWizard() {
	log.Info("start online wizard.")
	participates, err := self.participates.GetParticipates()
	if err != nil {
		log.Error("get participates failed with error %s.", err)
		self.notify()
		return
	}
	_, master, err := self.role.RoleAssignments(participates)
	if nil != err {
		log.Error("Role assignments failed with err %v.", err)
		self.notify()
		return
	}
	self.consensus.Initialization(master, participates, self.eventCenter, true)
	self.consensus.Online()
}

func (self *Node) mainLoop() {
	self.OnlineWizard()
	for {
		msg := <-self.msgChannel
		switch msg {
		case common.MsgBlockCommitSuccess:
			log.Info("Receive msg from switch is success.")
			self.NextRound(common.MsgBlockCommitSuccess)
		case common.MsgBlockCommitFailed:
			self.NextRound(common.MsgBlockCommitFailed)
			log.Info("Receive msg from switch is commit failed.")
		case common.MsgBlockVerifyFailed:
			log.Info("Receive msg from switch is verify failed.")
			self.NextRound(common.MsgBlockVerifyFailed)
		case common.MsgRoundRunFailed:
			log.Error("Receive msg from main loop is run failed.")
			self.NextRound(common.MsgBlockVerifyFailed)
		case common.MsgToConsensusFailed:
			log.Error("Receive msg of to consensus failed.")
			self.NextRound(common.MsgToConsensusFailed)
		case common.MsgChangeMaster:
			log.Info("Receive msg of change views.")
			self.NextRound(common.MsgBlockCommitSuccess)
		case common.MsgOnline:
			log.Info("Receive msg of online.")
			self.NextRound(common.MsgOnline)
		case common.MsgNodeServiceStopped:
			log.Warn("Stop node service.")
			break
		}
	}
}

func (self *Node) stratRpc() {
	var err error
	if self.rpcListeners, err = apigateway.StartRPC(self.config.ApiGatewayAddr); nil != err {
		log.Error("Start rpc failed with error %v.", err)
		panic("Rpc start failed.")
	}
}

func (self *Node) startSwitch() {
	if err := self.txSwitch.Start(); nil != err {
		log.Error("Start txs witch failed with error %v.", err)
		panic("TxSwitch start failed.")
	}
	if err := self.blockSwitch.Start(); nil != err {
		log.Error("Start block switch failed with error %v.", err)
		panic("BlockSwitch start failed.")
	}
}

func (self *Node) startBlockSyncer() {
	if err := self.blockSyncerP2P.Start(); nil != err {
		log.Error("Start block syncer p2p failed with error %v.", err)
		panic("Start block syncer p2p failed.")
	}
	if err := self.blockSyncer.Start(); nil != err {
		log.Error("Start block syncer failed with error %v.", err)
		panic("Start block syncer failed.")
	}
}

func (self *Node) startBlockPropagator() {
	if err := self.blockP2P.Start(); nil != err {
		log.Error("Start block p2p failed with error %v.", err)
		panic("Start block p2p failed.")
	}
	if err := self.blockPropagator.Start(); nil != err {
		log.Error("Start block propagator failed with error %v.", err)
		panic("Start block propagator failed.")
	}
}

func (self *Node) startTxPropagator() {
	if err := self.txP2P.Start(); nil != err {
		log.Error("Start tx p2p failed with error %v.", err)
		panic("Start tx p2p failed.")
	}
	if err := self.txPropagator.Start(); nil != err {
		log.Error("Start tx propagator failed with error %v.", err)
		panic("Start tx propagator failed.")
	}
}

func (self *Node) Start() {
	self.stratRpc()
	self.startSwitch()
	self.startBlockSyncer()
	self.startBlockPropagator()
	self.startTxPropagator()
	monitor.StartPrometheusServer(self.config.PrometheusConf)
	monitor.StartExpvarServer(self.config.ExpvarConf)
	monitor.StartPprofServer(self.config.PprofConf)
	if self.config.NodeType == common.ConsensusNode {
		go self.consensus.Start()
		go self.mainLoop()
	}
}

func (self *Node) Stop() error {
	log.Warn("Stop node service.")
	close(self.serviceChannel)
	for _, listener := range self.rpcListeners {
		if err := listener.Close(); err != nil {
			log.Error("Stop rpc listeners failed with error %v.", err)
			return fmt.Errorf("closing listener error")
		}
	}
	self.blockSyncerP2P.Stop()
	self.blockSyncer.Stop()
	self.blockP2P.Stop()
	self.blockPropagator.Stop()
	self.txP2P.Stop()
	self.txPropagator.Stop()
	self.blockSwitch.Stop()
	self.txSwitch.Stop()
	self.eventUnregister()
	if self.config.NodeType == common.ConsensusNode {
		self.msgChannel <- common.MsgNodeServiceStopped
		monitor.StopPrometheusServer()
	}
	return nil
}

func (self *Node) Wait() {
	<-self.serviceChannel
}

func (self *Node) Restart() error {
	if err := self.Stop(); err != nil {
		log.Error("restart service failed with err %v.", err)
		return err
	}
	self.Start()
	return nil
}
