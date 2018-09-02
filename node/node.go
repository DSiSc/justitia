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
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/gossipswitch"
	"github.com/DSiSc/justitia/config"
	"github.com/DSiSc/producer"
	"github.com/DSiSc/txpool"
	"github.com/DSiSc/txpool/log"
	"github.com/DSiSc/validator"
	"sync"
	"time"
)

var complete chan int

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
	complete = make(chan int)
	nodeConf := config.NewNodeConfig()

	txSwitch, err := gossipswitch.NewGossipSwitchByType(gossipswitch.TxSwitch)
	if err != nil {
		log.Error("Init txSwitch failed.")
		return nil, fmt.Errorf("TxSwitch failed.")
	}
	blkSwitch, err := gossipswitch.NewGossipSwitchByType(gossipswitch.BlockSwitch)
	if err != nil {
		log.Error("Init block switch failed.")
		return nil, fmt.Errorf("BlkSwitch failed.")
	}
	txpool := txpool.NewTxPool(nodeConf.TxPoolConf)
	swChIn := txSwitch.InPort(gossipswitch.LocalInPortId).Channel()
	rpc.SetSwCh(swChIn)
	err = txSwitch.OutPort(gossipswitch.LocalInPortId).BindToPort(func(msg gossipswitch.SwitchMsg) error {
		switch tx := msg.(type) {
		case *types.Transaction:
			return txpool.AddTx(tx)
		default:
			return fmt.Errorf("Not support msg type.")
		}
	})
	if err != nil {
		log.Error("Registe txpool failed.")
		return nil, fmt.Errorf("Registe txpool failed.")
	}

	err = blockchain.InitBlockChain(nodeConf.BlockChainConf)
	if err != nil {
		log.Error("Init blockchain failed.")
		return nil, fmt.Errorf("Blockchain failed.")
	}

	participates, err := participates.NewParticipates(nodeConf.ParticipatesConf)
	if nil != err {
		log.Error("Init participates failed.")
		return nil, fmt.Errorf("Participates failed.")
	}

	role, err := role.NewRole(participates, *nodeConf.Account, nodeConf.RoleConf)
	if nil != err {
		log.Error("Init role failed.")
		return nil, fmt.Errorf("Role failed.")
	}

	consensus, err := consensus.NewConsensus(participates, nodeConf.ConsensusConf)
	if nil != err {
		log.Error("Init consensus failed.")
		return nil, fmt.Errorf("Consensus failed.")
	}
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

func (self *Node) Round() {
	for {
		select {
		case <-complete:
			log.Warn("Stop node service.")
			self.nodeWg.Done()
			return
		default:
			// Waiting time in consistent.
			time.Sleep(10 * time.Nanosecond)
			log.Info("begin produce block.")
			// get role
			assigments, err := self.role.RoleAssignments()
			if nil != err {
				log.Error("Role assignments failed.")
				self.nodeWg.Done()
				continue
			}
			// new object based role
			if common.Master == assigments[*self.config.Account] {
				log.Info("I am master this round.")
				if nil == self.producer {
					self.producer = producer.NewProducer(self.txpool, self.config.Account)
				}
				// make block
				block, err := self.producer.MakeBlock()
				if err != nil {
					log.Error("Make block failed.")
					continue
				}
				// to consensus
				proposal := &consensusc.Proposal{
					Block: block,
				}
				err = self.consensus.ToConsensus(proposal)
				if err != nil {
					log.Error("Not to consensus.")
					continue
				}
				// broadcast block
				swChIn := self.blockSwitch.InPort(gossipswitch.LocalInPortId).Channel()
				swChIn <- proposal.Block
			}
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

	go self.Round()
	self.nodeWg.Wait()
	log.Warn("End start.")
}

func (self *Node) Stop() {
	log.Warn("Set node service stop.")
	complete <- 1
	return
}
