package node

import (
	"fmt"
	"github.com/DSiSc/apigateway"
	rpc "github.com/DSiSc/apigateway/rpc/core"
	"github.com/DSiSc/blockchain"
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
	"github.com/DSiSc/txpool/log"
	"github.com/DSiSc/validator"
	"sync"
	"sync/atomic"
)

var Complete atomic.Value

var MsgChannel = make(chan common.MsgType)

var (
	blockCommittedSub       types.Subscriber
	blockVerifyFaileddSub   types.Subscriber
	blockCommittedFailedSub types.Subscriber
)

type NodeService interface {
	Start()
	Stop()
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
}

func NewNode() (NodeService, error) {
	nodeConf := config.NewNodeConfig()
	types.GlobalEventCenter = events.NewEvent()

	txpool := txpool.NewTxPool(nodeConf.TxPoolConf)

	txSwitch, err := gossipswitch.NewGossipSwitchByType(gossipswitch.TxSwitch)
	if err != nil {
		log.Error("Init txSwitch failed.")
		return nil, fmt.Errorf("TxSwitch init failed.")
	}
	swChIn := txSwitch.InPort(gossipswitch.LocalInPortId).Channel()
	rpc.SetSwCh(swChIn)
	err = txSwitch.OutPort(gossipswitch.LocalInPortId).BindToPort(func(msg interface{}) error {
		return txpool.AddTx(msg.(*types.Transaction))
	})
	if err != nil {
		log.Error("Registe txpool failed.")
		return nil, fmt.Errorf("Registe txpool failed.")
	}

	blkSwitch, err := gossipswitch.NewGossipSwitchByType(gossipswitch.BlockSwitch)
	if err != nil {
		log.Error("Init block switch failed.")
		return nil, fmt.Errorf("BlkSwitch init failed.")
	}

	err = blockchain.InitBlockChain(nodeConf.BlockChainConf)
	if err != nil {
		log.Error("Init blockchain failed.")
		return nil, fmt.Errorf("Blockchain init failed.")
	}

	participates, err := participates.NewParticipates(nodeConf.ParticipatesConf)
	if nil != err {
		log.Error("Init participates failed.")
		return nil, fmt.Errorf("Participates init failed.")
	}

	role, err := role.NewRole(participates, *nodeConf.Account, nodeConf.RoleConf)
	if nil != err {
		log.Error("Init role failed.")
		return nil, fmt.Errorf("Role init failed.")
	}

	consensus, err := consensus.NewConsensus(participates, nodeConf.ConsensusConf)
	if nil != err {
		log.Error("Init consensus failed.")
		return nil, fmt.Errorf("Consensus init failed.")
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

func EventRegister() {
	blockCommittedSub = types.GlobalEventCenter.Subscribe(types.EventBlockCommitted, func(v interface{}) {
		MsgChannel <- common.MsgBlockCommitSuccess
	})
	blockVerifyFaileddSub = types.GlobalEventCenter.Subscribe(types.EventBlockVerifyFailed, func(v interface{}) {
		MsgChannel <- common.MsgBlockVerifyFailed
	})
	blockCommittedFailedSub = types.GlobalEventCenter.Subscribe(types.EventBlockCommitFailed, func(v interface{}) {
		MsgChannel <- common.MsgBlockCommitFailed
	})
}

func EventUnregister() {
	types.GlobalEventCenter.UnSubscribeAll()
}

func (self *Node) Round() error {
	log.Info("begin produce block.")
	assigments, err := self.role.RoleAssignments()
	if nil != err {
		log.Error("Role assignments failed.")
		return fmt.Errorf("Role assignments failed.")
	}
	if rolec.Master == assigments[*self.config.Account] {
		log.Info("I am master this round.")
		if nil == self.producer {
			self.producer = producer.NewProducer(self.txpool, self.config.Account)
		}
		block, err := self.producer.MakeBlock()
		if err != nil {
			log.Error("Make block failed.")
			return fmt.Errorf("Make block failed.")
		}
		proposal := &consensusc.Proposal{
			Block: block,
		}
		err = self.consensus.ToConsensus(proposal)
		if err != nil {
			log.Error("Not to consensus.")
			return fmt.Errorf("Not to consensus.")
		}
		swChIn := self.blockSwitch.InPort(gossipswitch.LocalInPortId).Channel()
		swChIn <- proposal.Block
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
		if Complete.Load() == true {
			log.Info("Stop node service.")
			break
		}
		err := self.Round()
		if nil != err {
			// if block make failed, then start a new round
			continue
		}
		msg := <-MsgChannel
		switch msg {
		case common.MsgBlockCommitSuccess:
		case common.MsgBlockCommitFailed:
		case common.MsgBlockVerifyFailed:
			continue
		default:
			log.Info("Not support msg type.")
			continue
		}
	}
}

func (self *Node) Start() {
	self.nodeWg.Add(1)
	err := self.txSwitch.Start()
	if nil != err {
		panic("TxSwitch Start Failed.")
	}
	err = self.blockSwitch.Start()
	if nil != err {
		panic("TxSwitch Start Failed.")
	}
	_, err = apigateway.StartRPC(self.config.ApiGatewayAddr)
	if nil != err {
		panic("Start rpc failed.")
	}
	Complete.Store(false)
	log.Info("start loop")
	go self.mainLoop()
	self.nodeWg.Wait()
}

func (self *Node) Stop() {
	log.Warn("Stop node service.")
	Complete.Store(true)
	EventUnregister()
	return
}
