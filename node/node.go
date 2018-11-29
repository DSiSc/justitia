package node

import (
	"fmt"
	"github.com/DSiSc/apigateway"
	rpc "github.com/DSiSc/apigateway/rpc/core"
	"github.com/DSiSc/blockchain"
	gconf "github.com/DSiSc/craft/config"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/monitor"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus"
	consensusc "github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role"
	commonr "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/gossipswitch"
	"github.com/DSiSc/justitia/common"
	"github.com/DSiSc/justitia/config"
	"github.com/DSiSc/justitia/tools/events"
	"github.com/DSiSc/producer"
	"github.com/DSiSc/txpool"
	"github.com/DSiSc/validator"
	"net"
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
	nodeWg         sync.WaitGroup
	config         config.NodeConfig
	txpool         txpool.TxsPool
	participates   participates.Participates
	role           role.Role
	consensus      consensus.Consensus
	producer       *producer.Producer
	txSwitch       *gossipswitch.GossipSwitch
	blockSwitch    *gossipswitch.GossipSwitch
	validator      *validator.Validator
	rpcListeners   []net.Listener
	eventCenter    types.EventCenter
	msgChannel     chan common.MsgType
	serviceChannel chan interface{}
}

func InitLog(args common.SysConfig, conf config.NodeConfig) {
	var logPath = args.LogPath
	if common.BlankString == logPath {
		logPath = conf.Logger.Output
	}
	var logFormat = args.LogStyle
	if common.BlankString == logFormat {
		logFormat = conf.Logger.Format
	}
	var logLevel = args.LogLevel
	if common.InvalidInt == int(logLevel) {
		logLevel = log.Level(conf.Logger.LogLevel)
	}

	log.SetTimestampFormat(conf.Logger.TimeFieldFormat)
	log.AddFileAppender(
		"filelog",
		logPath,
		logLevel,
		logFormat,
		conf.Logger.ShowCaller,
		conf.Logger.ShowHostname)
}

func NewNode(args common.SysConfig) (NodeService, error) {
	nodeConf := config.NewNodeConfig()
	InitLog(args, nodeConf)
	// record global hash algorithm
	gconf.GlobalConfig.Store(gconf.HashAlgName, nodeConf.AlgorithmConf.HashAlgorithm)
	pool := txpool.NewTxPool(nodeConf.TxPoolConf)
	eventsCenter := events.NewEvent()
	txSwitch, err := gossipswitch.NewGossipSwitchByType(gossipswitch.TxSwitch, eventsCenter)
	if err != nil {
		log.Error("Init txSwitch failed.")
		return nil, fmt.Errorf("txswitch init failed")
	}
	swChIn := txSwitch.InPort(gossipswitch.LocalInPortId).Channel()
	rpc.SetSwCh(swChIn)
	err = txSwitch.OutPort(gossipswitch.LocalInPortId).BindToPort(func(msg interface{}) error {
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

	participates, err := participates.NewParticipates(nodeConf.ParticipatesConf)
	if nil != err {
		log.Error("Init participates failed.")
		return nil, fmt.Errorf("participates init failed")
	}

	role, err := role.NewRole(nodeConf.RoleConf)
	if nil != err {
		log.Error("Init role failed.")
		return nil, fmt.Errorf("role init failed")
	}

	consensus, err := consensus.NewConsensus(participates, nodeConf.ConsensusConf, nodeConf.Account)
	if nil != err {
		log.Error("Init consensus failed.")
		return nil, fmt.Errorf("consensus init failed")
	}
	node := &Node{
		config:         nodeConf,
		txpool:         pool,
		participates:   participates,
		role:           role,
		consensus:      consensus,
		txSwitch:       txSwitch,
		blockSwitch:    blkSwitch,
		eventCenter:    eventsCenter,
		msgChannel:     make(chan common.MsgType),
		serviceChannel: make(chan interface{}),
	}
	node.eventsRegister()
	return node, nil
}

func (self *Node) eventsRegister() {
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
}

func (self *Node) eventUnregister() {
	self.eventCenter.UnSubscribeAll()
}

func (self *Node) notify() {
	go func() {
		self.msgChannel <- common.MsgRoundRunFailed
	}()
}

func (self *Node) Round() {
	log.Debug("start a new round.")
	time.Sleep(time.Duration(self.config.BlockInterval) * time.Second)
	participates, err := self.participates.GetParticipates()
	if err != nil {
		log.Error("get participates failed with error %s.", err)
		self.notify()
		return
	}
	assignments, err := self.role.RoleAssignments(participates)
	if nil != err {
		log.Error("Role assignments failed with err %v.", err)
		self.notify()
		return
	}
	self.consensus.Initialization(assignments, participates, self.eventCenter)
	role, ok := assignments[self.config.Account]
	if ok && (commonr.Master == role) {
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
		proposal := &consensusc.Proposal{
			Block: block,
		}
		if err = self.consensus.ToConsensus(proposal); err != nil {
			log.Error("ToConsensus failed with err %v.", err)
		} else {
			block.HeaderHash = common.HeaderHash(block)
			self.txpool.DelTxs(block.Transactions)
			// TODO: notify p2p to broadcast block
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

func (self *Node) mainLoop() {
	for {
		self.Round()
		msg := <-self.msgChannel
		switch msg {
		case common.MsgBlockCommitSuccess:
			log.Info("Receive msg from switch is success.")
		case common.MsgBlockCommitFailed:
			log.Info("Receive msg from switch is commit failed.")
		case common.MsgBlockVerifyFailed:
			log.Info("Receive msg from switch is verify failed.")
		case common.MsgRoundRunFailed:
			log.Error("Receive msg from main loop is run failed.")
		case common.MsgToConsensusFailed:
			log.Error("Receive msg of to consensus failed.")
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

func (self *Node) Start() {
	self.stratRpc()
	self.startSwitch()
	monitor.StartPrometheusServer(self.config.PrometheusConf)
	monitor.StartExpvarServer(self.config.ExpvarConf)
	go self.consensus.Start()
	go self.mainLoop()
}

func (self *Node) Stop() error {
	log.Warn("Stop node service.")
	close(self.serviceChannel)
	self.msgChannel <- common.MsgNodeServiceStopped
	for _, listener := range self.rpcListeners {
		if err := listener.Close(); err != nil {
			log.Error("Stop rpc listeners failed with error %v.", err)
			return fmt.Errorf("closing listener error")
		}
	}
	monitor.StopPrometheusServer()
	self.blockSwitch.Stop()
	self.txSwitch.Stop()
	self.eventUnregister()
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
