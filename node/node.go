package node

import (
	"fmt"
	"github.com/DSiSc/galaxy/consensus"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role"
	"github.com/DSiSc/galaxy/role/common"
	"github.com/DSiSc/gossipswitch"
	"github.com/DSiSc/justitia/config"
	"github.com/DSiSc/ledger"
	"github.com/DSiSc/producer"
	"github.com/DSiSc/txpool"
	"github.com/DSiSc/txpool/log"
)

type NodeService interface {
	Start() error
}

// node struct with all service
type Node struct {
	config       config.NodeConfig
	txpool       txpool.TxsPool
	participates participates.Participates
	role         role.Role
	consensus    consensus.Consensus
	ledger       *ledger.Ledger
	producer     *producer.Producer
	txSwitch     *gossipswitch.GossipSwitch
	blockSwitch  *gossipswitch.GossipSwitch
}

func NewNode() (NodeService, error) {
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

	ledger, err := ledger.NewLedger(nodeConf.LedgerConf)
	if nil != err {
		log.Error("Init leger store failed.")
		return nil, fmt.Errorf("Ledger store failed.")
	}

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
		ledger:       ledger,
		txSwitch:     txSwitch,
		blockSwitch:  blkSwitch,
	}

	return node, nil
}

func (self *Node) Start() error {
	// TODO: execute it every round
	assigments, err := self.role.RoleAssignments()
	if nil != err {
		log.Error("Role assignments failed.")
		return fmt.Errorf("Role assignments failed.")
	}

	if common.Master == assigments[self.config.Account] {
		if nil == self.producer {
			producer, err1 := producer.NewProducer(self.txpool, self.ledger)
			if nil != err1 {
				log.Error("New producer failed.")
				return fmt.Errorf("Init producer failed.")
			}
			self.producer = producer
		}
		_, err2 := self.producer.MakeBlock()
		if nil != err2 {
			log.Error("Make block failed.")
			return fmt.Errorf("Make block failed.")
		}
	}
	return nil
}
