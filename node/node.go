package node

import (
	"fmt"
	"github.com/DSiSc/galaxy/consensus"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/role"
	"github.com/DSiSc/justitia/node/config"
	"github.com/DSiSc/txpool/common/log"
	"github.com/DSiSc/txpool/core"
)

// node struct with all service
type Node struct {
	config       config.NodeConfig
	txpool       core.TxsPool
	participates participates.Participates
	role         role.Role
	consensus    consensus.Consensus
}

// init a node fimply
func NewNode() (*Node, error) {

	nodeConfig := config.NewNodeConfig()

	// txpool
	txpoolConf := core.TxPoolConfig{
		GlobalSlots: nodeConfig.GlobalSlots,
	}
	txpool := core.NewTxPool(txpoolConf)

	// participate
	participates, err := participates.NewParticipatePolicy()
	if nil != err {
		log.Error("Init participate failed.")
		return nil, fmt.Errorf("Participates failed.")
	}

	// role
	role, err := role.NewRolePolicy(participates, nodeConfig.Account)
	if nil != err {
		log.Error("Init role failed.")
		return nil, fmt.Errorf("Role failed.")
	}

	// consensus
	consensus, err := consensus.NewConsensusPolicy(participates)
	if nil != err {
		log.Error("Init consensus failed.")
		return nil, fmt.Errorf("Consensus failed.")
	}

	node := &Node{
		config: nodeConfig,
		txpool: txpool,
		participates: participates,
		role: role,
		consensus: consensus,
	}

	return node, nil
}
