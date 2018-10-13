package node

import (
	"fmt"
	"net"
	"sync"
	"time"

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
	rolec "github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/gossipswitch"
	"github.com/DSiSc/justitia/common"
	"github.com/DSiSc/justitia/config"
	"github.com/DSiSc/justitia/tools/events"
	"github.com/DSiSc/producer"
	"github.com/DSiSc/txpool"
	"github.com/DSiSc/validator"
)

var (
	MsgChannel chan common.MsgType
	StopSignal chan interface{}
)

type NodeService interface {
	Start()
	Stop() error
	Wait()
	Restart() error
}

// node struct with all service
type Node struct {
	nodeWg       sync.WaitGroup
	config       config.NodeConfig
	txpool       txpool.TxsPool
	participates participates.Participates
	role         role.Role
	consensus    consensus.Consensus
	producer     *producer.Producer
	txSwitch     *gossipswitch.GossipSwitch
	blockSwitch  *gossipswitch.GossipSwitch
	validator    *validator.Validator
	rpcListeners []net.Listener
}

func EventRegister() {
	types.GlobalEventCenter.Subscribe(types.EventBlockCommitted, func(v interface{}) {
		MsgChannel <- common.MsgBlockCommitSuccess
	})
	types.GlobalEventCenter.Subscribe(types.EventBlockVerifyFailed, func(v interface{}) {
		MsgChannel <- common.MsgBlockVerifyFailed
	})
	types.GlobalEventCenter.Subscribe(types.EventBlockCommitFailed, func(v interface{}) {
		MsgChannel <- common.MsgBlockCommitFailed
	})
}

func NewNode() (NodeService, error) {
	log.AddFileAppender(
		"filelog", "/tmp/DSiSc/justitia.log", log.InfoLevel, log.TextFmt, true, true)
	nodeConf := config.NewNodeConfig()
	// record global hash algorithm
	gconf.GlobalConfig.Store(gconf.HashAlgName, nodeConf.AlgorithmConf.HashAlgorithm)

	types.GlobalEventCenter = events.NewEvent()
	txpool := txpool.NewTxPool(nodeConf.TxPoolConf)
	MsgChannel = make(chan common.MsgType)
	StopSignal = make(chan interface{})
	txSwitch, err := gossipswitch.NewGossipSwitchByType(gossipswitch.TxSwitch)
	if err != nil {
		log.Error("Init txSwitch failed.")
		return nil, fmt.Errorf("txswitch init failed")
	}
	swChIn := txSwitch.InPort(gossipswitch.LocalInPortId).Channel()
	rpc.SetSwCh(swChIn)
	err = txSwitch.OutPort(gossipswitch.LocalInPortId).BindToPort(func(msg interface{}) error {
		return txpool.AddTx(msg.(*types.Transaction))
	})
	if err != nil {
		log.Error("Register txpool failed.")
		return nil, fmt.Errorf("registe txpool failed")
	}

	blkSwitch, err := gossipswitch.NewGossipSwitchByType(gossipswitch.BlockSwitch)
	if err != nil {
		log.Error("Init block switch failed.")
		return nil, fmt.Errorf("blkSwitch init failed")
	}

	err = blockchain.InitBlockChain(nodeConf.BlockChainConf)
	if err != nil {
		log.Error("Init blockchain failed.")
		return nil, fmt.Errorf("blockchain init failed")
	}

	participates, err := participates.NewParticipates(nodeConf.ParticipatesConf)
	if nil != err {
		log.Error("Init participates failed.")
		return nil, fmt.Errorf("participates init failed")
	}

	role, err := role.NewRole(participates, *nodeConf.Account, nodeConf.RoleConf)
	if nil != err {
		log.Error("Init role failed.")
		return nil, fmt.Errorf("role init failed")
	}

	consensus, err := consensus.NewConsensus(participates, nodeConf.ConsensusConf)
	if nil != err {
		log.Error("Init consensus failed.")
		return nil, fmt.Errorf("consensus init failed")
	}
	EventRegister()
	node := &Node{
		config:       nodeConf,
		txpool:       txpool,
		participates: participates,
		role:         role,
		consensus:    consensus,
		txSwitch:     txSwitch,
		blockSwitch:  blkSwitch,
	}

	return node, nil
}

func EventUnregister() {
	types.GlobalEventCenter.UnSubscribeAll()
}

func (self *Node) Round() error {
	time.Sleep(time.Duration(self.config.BlockInterval) * time.Second)
	log.Info("Begin to produce block.")
	assignments, err := self.role.RoleAssignments()
	if nil != err {
		log.Error("Role assignments failed.")
		return fmt.Errorf("role assignments failed")
	}
	if rolec.Master == assignments[*self.config.Account] {
		log.Info("I am master this round.")
		if nil == self.producer {
			self.producer = producer.NewProducer(self.txpool, self.config.Account)
		}
		block, err := self.producer.MakeBlock()
		if err != nil {
			log.Error("Make block failed.")
			return fmt.Errorf("make block failed")
		}
		proposal := &consensusc.Proposal{
			Block: block,
		}
		if err = self.consensus.ToConsensus(proposal); err != nil {
			log.Error("Not to consensus.")
			return fmt.Errorf("consensus failed")
		}
		block.HeaderHash = common.HeaderHash(block)
		swChIn := self.blockSwitch.InPort(gossipswitch.LocalInPortId).Channel()
		swChIn <- proposal.Block
		log.Info("New block height is: %d.", block.Header.Height)
	} else {
		if self.validator == nil {
			// TODO: attach validator to consensus
			self.validator = validator.NewValidator(self.config.Account)
		}
	}
	return nil
}

func (self *Node) mainLoop() {
	for {
		if err := self.Round(); nil != err {
			// if block make failed, then start a new round
			log.Error("This round failed with err %v.", nil)
			continue
		}
		msg := <-MsgChannel
		switch msg {
		case common.MsgBlockCommitSuccess:
			log.Info("Receive msg from switch is success.")
		case common.MsgBlockCommitFailed:
			log.Info("Receive msg from switch is commit failed.")
		case common.MsgBlockVerifyFailed:
			log.Info("Receive msg from switch is verify failed.")
		case common.MsgNodeServiceStopped:
			return
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
		log.Error("Start txswitch failed with error %v.", err)
		panic("TxSwitch start failed.")
	}
	if err := self.blockSwitch.Start(); nil != err {
		log.Error("Start blockswitch failed with error %v.", err)
		panic("BlockSwitch start failed.")
	}
}

func (self *Node) Start() {
	self.stratRpc()
	self.startSwitch()

	monitor.StartPrometheusServer(self.config.PrometheusConf)

	go self.mainLoop()
}

func (self *Node) Stop() error {
	log.Warn("Stop node service.")
	close(StopSignal)
	MsgChannel <- common.MsgNodeServiceStopped
	for _, l := range self.rpcListeners {
		log.Info("Closing rpc listener")
		if err := l.Close(); err != nil {
			log.Error("Stop rpc listeners failed with error %v.", err)
			return fmt.Errorf("closing listener error")
		}
	}

	monitor.StopPrometheusServer()

	self.blockSwitch.Stop()
	self.txSwitch.Stop()
	EventUnregister()
	return nil
}

func (self *Node) Wait() {
	<-StopSignal
}

func (self *Node) Restart() error {
	if err := self.Stop(); err != nil {
		return err
	}
	self.Start()
	return nil
}
