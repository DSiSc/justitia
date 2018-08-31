package node

import (
	"fmt"
	"github.com/DSiSc/apigateway"
	rpc "github.com/DSiSc/apigateway/rpc/core"
	"github.com/DSiSc/blockchain"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus"
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
	account      *types.Account
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

	swCh := txSwitch.InPort(gossipswitch.LocalInPortId).Channel()
	rpc.SetSwCh(swCh)

	err = blockchain.InitBlockChain(nodeConf.BlockChainConf)
	if err != nil {
		log.Error("Init blockchain failed.")
		return nil, fmt.Errorf("Blockchain failed.")
	}

	txpool := txpool.NewTxPool(nodeConf.TxPoolConf)
	participates, err := participates.NewParticipates(nodeConf.ParticipatesConf)
	if nil != err {
		log.Error("Init participates failed.")
		return nil, fmt.Errorf("Participates failed.")
	}

	role, err := role.NewRole(participates, nodeConf.Account, nodeConf.RoleConf)
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
				return
			}
			// new object based role
			if common.Master == assigments[self.config.Account] {
				log.Info("I am master this round.")
				if nil == self.producer {
					producer, err1 := producer.NewProducer(self.txpool, nil)
					if nil != err1 {
						log.Error("New producer failed.")
						self.nodeWg.Done()
						return
					}
					self.producer = producer
				} else {
					log.Info("I am slave this round.")
					if nil == self.validator {
						self.validator = validator.NewValidator(self.account)
					}
				}
			}
		}
	}
}

func (self *Node) Start() {
	self.nodeWg.Add(1)
	_, err := apigateway.StartRPC(self.config.ApiGatewayAddr)
	if nil != err {
		log.Warn("Start rpc failed.")
		return
	}
	// TODO: start rpc service
	go self.Round()
	self.nodeWg.Wait()
	log.Warn("End start.")
}

func (self *Node) Stop() {
	log.Warn("Set node service stop.")
	complete <- 1
	return
}
