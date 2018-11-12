package node

import (
	"fmt"
	"github.com/DSiSc/apigateway"
	"github.com/DSiSc/blockchain"
	blockchainc "github.com/DSiSc/blockchain/config"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus"
	gcommon "github.com/DSiSc/galaxy/consensus/common"
	consensusc "github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/policy/solo"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/role"
	"github.com/DSiSc/galaxy/role/common"
	rolec "github.com/DSiSc/galaxy/role/config"
	solo2 "github.com/DSiSc/galaxy/role/policy/solo"
	"github.com/DSiSc/gossipswitch"
	commonc "github.com/DSiSc/justitia/common"
	"github.com/DSiSc/justitia/tools/events"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/producer"
	"github.com/DSiSc/validator/tools/account"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"reflect"
	"testing"
)

var defaultConf = commonc.SysConfig{
	LogLevel: log.InfoLevel,
	LogPath:  "/tmp/justitia.log",
	LogStyle: "json",
}

func TestNewNode(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(gossipswitch.NewGossipSwitchByType, func(switchType gossipswitch.SwitchType) (*gossipswitch.GossipSwitch, error) {
		return nil, fmt.Errorf("mock gossipswitch error")
	})
	monkey.Patch(log.AddAppender, func(appenderName string, output io.Writer, logLevel log.Level, format string, showCaller bool, showHostname bool) {
		return
	})
	service, err := NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("txswitch init failed"))
	monkey.Unpatch(gossipswitch.NewGossipSwitchByType)

	monkey.Patch(gossipswitch.NewGossipSwitchByType, func(switchType gossipswitch.SwitchType) (*gossipswitch.GossipSwitch, error) {
		return nil, fmt.Errorf("mock gossipswitch error")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("txswitch init failed"))
	monkey.Unpatch(gossipswitch.NewGossipSwitchByType)

	var op *gossipswitch.OutPort
	monkey.PatchInstanceMethod(reflect.TypeOf(op), "BindToPort", func(_ *gossipswitch.OutPort, _ gossipswitch.OutPutFunc) error {
		return fmt.Errorf("bind error")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("registe txpool failed"))
	monkey.PatchInstanceMethod(reflect.TypeOf(op), "BindToPort", func(_ *gossipswitch.OutPort, _ gossipswitch.OutPutFunc) error {
		return nil
	})

	monkey.Patch(blockchain.InitBlockChain, func(_ blockchainc.BlockChainConfig) error {
		return fmt.Errorf("mock blockchain error")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("blockchain init failed"))
	monkey.Unpatch(blockchain.InitBlockChain)

	monkey.Patch(participates.NewParticipates, func(conf config.ParticipateConfig) (participates.Participates, error) {
		return nil, fmt.Errorf("mock participates error")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("participates init failed"))
	monkey.Unpatch(participates.NewParticipates)

	monkey.Patch(role.NewRole, func(participates.Participates, rolec.RoleConfig) (role.Role, error) {
		return nil, fmt.Errorf("mock role error")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("role init failed"))
	monkey.Unpatch(role.NewRole)

	monkey.Patch(consensus.NewConsensus, func(participates.Participates, consensusc.ConsensusConfig, account.Account) (consensus.Consensus, error) {
		return nil, fmt.Errorf("mock consensus error")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("consensus init failed"))
	monkey.Unpatch(consensus.NewConsensus)

	service, err = NewNode(defaultConf)
	assert.Nil(err)
	assert.NotNil(service)

	nodeService := service.(*Node)
	assert.NotNil(nodeService.txpool)
	assert.NotNil(nodeService.participates)
	assert.NotNil(nodeService.txSwitch)
	assert.NotNil(nodeService.blockSwitch)
	assert.NotNil(nodeService.consensus)
	assert.NotNil(nodeService.config)
	assert.NotNil(nodeService.role)
	assert.Nil(nodeService.producer)
	assert.Nil(nodeService.validator)
	event := types.GlobalEventCenter.(*events.Event)
	assert.Equal(3, len(event.Subscribers))
}

func TestNode_Start(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(log.AddAppender, func(appenderName string, output io.Writer, logLevel log.Level, format string, showCaller bool, showHostname bool) {
		return
	})
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	assert.NotNil(service)
	go func() {
		monkey.Patch(apigateway.StartRPC, func(string) ([]net.Listener, error) {
			return make([]net.Listener, 0), nil
		})
		var c *solo.SoloPolicy
		monkey.PatchInstanceMethod(reflect.TypeOf(c), "Start", func(*solo.SoloPolicy) {
			return
		})
		service.Start()
		nodeService := service.(*Node)
		assert.NotNil(nodeService.rpcListeners)
		assert.Equal(0, len(nodeService.rpcListeners))
	}()
	service.Stop()
}

func TestEventRegister(t *testing.T) {
	assert := assert.New(t)
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	assert.NotNil(service)
	node := service.(*Node)
	types.GlobalEventCenter = events.NewEvent()
	EventRegister(node)
	event := types.GlobalEventCenter.(*events.Event)
	assert.Equal(3, len(event.Subscribers))
}

func TestEventUnregister(t *testing.T) {
	assert := assert.New(t)
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	assert.NotNil(service)
	node := service.(*Node)
	// init event center
	types.GlobalEventCenter = events.NewEvent()
	EventRegister(node)
	eventC := types.GlobalEventCenter.(*events.Event)
	assert.Equal(3, len(eventC.Subscribers))
}

func TestNode_Restart(t *testing.T) {
	var nodeService *Node
	var node *Node
	monkey.PatchInstanceMethod(reflect.TypeOf(node), "Stop", func(_ *Node) error {
		return fmt.Errorf("node stop error")
	})
	err := nodeService.Restart()
	assert.Equal(t, err, fmt.Errorf("node stop error"))
	monkey.UnpatchInstanceMethod(reflect.TypeOf(node), "Stop")

	monkey.PatchInstanceMethod(reflect.TypeOf(node), "Stop", func(*Node) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(node), "Start", func(*Node) {
		return
	})
	err = nodeService.Restart()
	assert.Nil(t, err)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(node), "Stop")
	monkey.UnpatchInstanceMethod(reflect.TypeOf(node), "Start")
}

var mockAccount = account.Account{
	Address: types.Address{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
		0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
	Extension: account.AccountExtension{
		Id:  0,
		Url: "172.0.0.1:8080",
	},
}

func TestNode_Round(t *testing.T) {
	assert := assert.New(t)
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	assert.NotNil(service)
	node := service.(*Node)

	var r *solo2.SoloPolicy
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "RoleAssignments", func(*solo2.SoloPolicy) (map[account.Account]common.Roler, error) {
		return nil, fmt.Errorf("assignments failed")
	})
	node.Round()

	monkey.PatchInstanceMethod(reflect.TypeOf(r), "RoleAssignments", func(*solo2.SoloPolicy) (map[account.Account]common.Roler, error) {
		role := make(map[account.Account]common.Roler)
		role[node.config.Account] = common.Slave
		return role, nil
	})
	assert.Nil(node.validator)
	node.Round()

	monkey.PatchInstanceMethod(reflect.TypeOf(r), "RoleAssignments", func(*solo2.SoloPolicy) (map[account.Account]common.Roler, error) {
		role := make(map[account.Account]common.Roler)
		role[node.config.Account] = common.Master
		return role, nil
	})
	var p *producer.Producer
	monkey.PatchInstanceMethod(reflect.TypeOf(p), "MakeBlock", func(*producer.Producer) (*types.Block, error) {
		return &types.Block{}, fmt.Errorf("make block failed")
	})
	assert.Nil(node.producer)
	node.Round()

	monkey.PatchInstanceMethod(reflect.TypeOf(p), "MakeBlock", func(*producer.Producer) (*types.Block, error) {
		return &types.Block{}, nil
	})
	var c *solo.SoloPolicy
	monkey.PatchInstanceMethod(reflect.TypeOf(c), "ToConsensus", func(*solo.SoloPolicy, *gcommon.Proposal) error {
		return fmt.Errorf("consensus failed")
	})
	node.Round()

}
