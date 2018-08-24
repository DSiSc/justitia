package node

import (
	"fmt"
	"github.com/DSiSc/galaxy/consensus"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role"
	"github.com/DSiSc/gossipswitch"
	"github.com/DSiSc/justitia/config"
	"github.com/DSiSc/ledger"
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
	txSwitch     *gossipswitch.GossipSwitch
	blockSwitch  *gossipswitch.GossipSwitch
}

// init a node fimply
func NewNode() (NodeService, error) {
	nodeConf := config.NewNodeConfig()
	// gossip switch
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
	// txpool
	txpool := txpool.NewTxPool(nodeConf.TxPoolConf)
	// ledger
	ledger, err := ledger.NewLedger(nodeConf.LedgerConf)
	if nil != err {
		log.Error("Init leger store failed.")
		return nil, fmt.Errorf("Ledger store failed.")
	}
	// participate
	participates, err := participates.NewParticipates(nodeConf.ParticipatesConf)
	if nil != err {
		log.Error("Init participates failed.")
		return nil, fmt.Errorf("Participates failed.")
	}
	// role
	role, err := role.NewRole(participates, nodeConf.Account, nodeConf.RoleConf)
	if nil != err {
		log.Error("Init role failed.")
		return nil, fmt.Errorf("Role failed.")
	}
	// consensus
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
	// start
	_, err := self.participates.GetParticipates()
	if nil != err {
		log.Error("Get participates failed.")
		return fmt.Errorf("Get participates failed.")
	}
	_, err = self.role.RoleAssignments()
	if nil != err {
		log.Error("Role assignments failed.")
		return fmt.Errorf("Role assignments failed.")
	}
	return nil
}
