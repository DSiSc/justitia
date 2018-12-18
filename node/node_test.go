package node

import (
	"fmt"
	"github.com/DSiSc/apigateway"
	"github.com/DSiSc/blockchain"
	blockchainc "github.com/DSiSc/blockchain/config"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy"
	galaxyCommon "github.com/DSiSc/galaxy/common"
	gcommon "github.com/DSiSc/galaxy/consensus/common"
	"github.com/DSiSc/galaxy/consensus/policy/dbft"
	"github.com/DSiSc/galaxy/consensus/policy/fbft"
	"github.com/DSiSc/galaxy/consensus/policy/solo"
	"github.com/DSiSc/galaxy/role/common"
	galaxySolo "github.com/DSiSc/galaxy/role/policy/solo"
	"github.com/DSiSc/gossipswitch"
	"github.com/DSiSc/gossipswitch/port"
	justitiaCommon "github.com/DSiSc/justitia/common"
	"github.com/DSiSc/justitia/config"
	"github.com/DSiSc/justitia/propagator"
	"github.com/DSiSc/justitia/tools/events"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/p2p"
	"github.com/DSiSc/producer"
	"github.com/DSiSc/syncer"
	"github.com/DSiSc/validator/tools/account"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"testing"
)

var defaultConf = justitiaCommon.SysConfig{
	LogLevel: log.InfoLevel,
	LogPath:  "/tmp/justitia.log",
	LogStyle: "json",
}

func TestNewNode(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(gossipswitch.NewGossipSwitchByType, func(switchType gossipswitch.SwitchType, _ types.EventCenter) (*gossipswitch.GossipSwitch, error) {
		if gossipswitch.TxSwitch == switchType {
			return nil, fmt.Errorf("mock gossipswitch error")
		}
		return nil, nil
	})
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	service, err := NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("txswitch init failed"))
	monkey.Unpatch(gossipswitch.NewGossipSwitchByType)

	var op *port.OutPort
	monkey.PatchInstanceMethod(reflect.TypeOf(op), "BindToPort", func(_ *port.OutPort, _ port.OutPutFunc) error {
		return fmt.Errorf("bind error")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("registe txpool failed"))
	monkey.UnpatchInstanceMethod(reflect.TypeOf(op), "BindToPort")

	monkey.PatchInstanceMethod(reflect.TypeOf(op), "BindToPort", func(_ *port.OutPort, _ port.OutPutFunc) error {
		return nil
	})
	monkey.Patch(blockchain.InitBlockChain, func(blockchainc.BlockChainConfig, types.EventCenter) error {
		return fmt.Errorf("mock blockchain error")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("blockchain init failed"))
	monkey.Unpatch(blockchain.InitBlockChain)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(op), "BindToPort")
	/*
		monkey.Patch(participates.NewParticipates, func(conf config.ParticipateConfig) (participates.Participates, error) {
			return nil, fmt.Errorf("mock participates error")
		})
		service, err = NewNode(defaultConf)
		assert.NotNil(err)
		assert.Nil(service)
		assert.Equal(err, fmt.Errorf("participates init failed"))
		monkey.Unpatch(participates.NewParticipates)

		monkey.Patch(role.NewRole, func(rolec.RoleConfig) (role.Role, error) {
			return nil, fmt.Errorf("mock role error")
		})
		service, err = NewNode(defaultConf)
		assert.NotNil(err)
		assert.Nil(service)
		assert.Equal(err, fmt.Errorf("role init failed"))
		monkey.Unpatch(role.NewRole)

		monkey.Patch(consensus.NewConsensus, func(participates.Participates, consensusc.ConsensusConfig, account.Account, chan<- interface{}) (consensus.Consensus, error) {
			return nil, fmt.Errorf("mock consensus error")
		})
		service, err = NewNode(defaultConf)
		assert.NotNil(err)
		assert.Nil(service)
		assert.Equal(err, fmt.Errorf("consensus init failed"))
		monkey.Unpatch(consensus.NewConsensus)
    */
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
	event := nodeService.eventCenter.(*events.Event)
	assert.Equal(6, len(event.Subscribers))

	monkey.Patch(galaxy.NewGalaxyPlugin, func(galaxyCommon.GalaxyPluginConf) (*galaxyCommon.GalaxyPlugin, error) {
		return nil, fmt.Errorf("error of NewGalaxyPlugin")
	})
	service, err = NewNode(defaultConf)
	assert.Equal(err, fmt.Errorf("init galaxy plugin failed with error error of NewGalaxyPlugin"))
	assert.NotNil(service)
	monkey.Unpatch(galaxy.NewGalaxyPlugin)

	nodeConf := config.NewNodeConfig()
	monkey.Patch(config.NewNodeConfig, func() config.NodeConfig {
		nodeConf.NodeType = justitiaCommon.FullNode
		return nodeConf
	})
	service, err = NewNode(defaultConf)
	nodeService = service.(*Node)
	event = nodeService.eventCenter.(*events.Event)
	assert.Equal(1, len(event.Subscribers))
	assert.Nil(nodeService.participates)
	assert.Nil(nodeService.consensus)
	assert.Nil(nodeService.role)
	monkey.Unpatch(config.NewNodeConfig)
	monkey.Unpatch(config.GetLogSetting)
	monkey.Unpatch(InitLog)
}

func TestNode_Start(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(log.Error, func(mtmsg string, a ...interface{}) {
		return
	})
	monkey.Patch(log.Info, func(mtmsg string, a ...interface{}) {
		return
	})
	monkey.Patch(log.Debug, func(mtmsg string, a ...interface{}) {
		return
	})
	monkey.Patch(log.Warn, func(mtmsg string, a ...interface{}) {
		return
	})
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	monkey.Patch(blockchain.InitBlockChain, func(blockchainc.BlockChainConfig, types.EventCenter) error {
		return nil
	})
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	assert.NotNil(service)
	var ch = make(chan int)
	go func() {
		monkey.Patch(log.Error, func(mtmsg string, a ...interface{}) {
			return
		})
		monkey.Patch(log.Info, func(mtmsg string, a ...interface{}) {
			return
		})
		monkey.Patch(log.Debug, func(mtmsg string, a ...interface{}) {
			return
		})
		monkey.Patch(log.Warn, func(mtmsg string, a ...interface{}) {
			return
		})
		monkey.Patch(log.SetGlobalConfig, func(config *log.Config) {
			return
		})
		monkey.Patch(apigateway.StartRPC, func(string) ([]net.Listener, error) {
			return make([]net.Listener, 0), nil
		})
		var c *solo.SoloPolicy
		monkey.PatchInstanceMethod(reflect.TypeOf(c), "Start", func(*solo.SoloPolicy) {
			return
		})
		var p *p2p.P2P
		monkey.PatchInstanceMethod(reflect.TypeOf(p), "Start", func(*p2p.P2P) error {
			return nil
		})
		monkey.PatchInstanceMethod(reflect.TypeOf(p), "Stop", func(*p2p.P2P) {
			return
		})
		var s *syncer.BlockSyncer
		monkey.PatchInstanceMethod(reflect.TypeOf(s), "Start", func(*syncer.BlockSyncer) error {
			return nil
		})
		monkey.PatchInstanceMethod(reflect.TypeOf(s), "Stop", func(*syncer.BlockSyncer) {
			return
		})
		var pb *propagator.BlockPropagator
		monkey.PatchInstanceMethod(reflect.TypeOf(pb), "Start", func(*propagator.BlockPropagator) error {
			return nil
		})
		monkey.PatchInstanceMethod(reflect.TypeOf(pb), "Stop", func(*propagator.BlockPropagator) {
			return
		})
		var pt *propagator.TxPropagator
		monkey.PatchInstanceMethod(reflect.TypeOf(pt), "Start", func(*propagator.TxPropagator) error {
			return nil
		})
		monkey.PatchInstanceMethod(reflect.TypeOf(pt), "Stop", func(*propagator.TxPropagator) {
			return
		})

		service.Start()
		ch <- 1
		nodeService := service.(*Node)
		assert.NotNil(nodeService.rpcListeners)
		assert.Equal(0, len(nodeService.rpcListeners))
		monkey.Unpatch(apigateway.StartRPC)
		monkey.UnpatchInstanceMethod(reflect.TypeOf(c), "Start")
		monkey.UnpatchInstanceMethod(reflect.TypeOf(p), "Start")
		monkey.UnpatchInstanceMethod(reflect.TypeOf(p), "Stop")
		monkey.UnpatchInstanceMethod(reflect.TypeOf(s), "Start")
		monkey.UnpatchInstanceMethod(reflect.TypeOf(s), "Stop")
		monkey.UnpatchInstanceMethod(reflect.TypeOf(pb), "Start")
		monkey.UnpatchInstanceMethod(reflect.TypeOf(pb), "Stop")
		monkey.UnpatchInstanceMethod(reflect.TypeOf(pt), "Start")
		monkey.UnpatchInstanceMethod(reflect.TypeOf(pt), "Stop")
		monkey.UnpatchAll()
	}()
	<-ch
	service.Stop()
	monkey.UnpatchAll()
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

var mockAccounts = []account.Account{
	account.Account{
		Address: types.Address{0x33, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  0,
			Url: "127.0.0.1:8080",
		},
	},
	account.Account{
		Address: types.Address{0x34, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68,
			0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  1,
			Url: "127.0.0.1:8081"},
	},
	account.Account{
		Address: types.Address{0x35, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  2,
			Url: "127.0.0.1:8082",
		},
	},

	account.Account{
		Address: types.Address{0x36, 0x3c, 0x33, 0x10, 0x82, 0x4b, 0x7c, 0x68, 0x51, 0x33, 0xf2, 0xbe, 0xdb, 0x2c, 0xa4, 0xb8, 0xb4, 0xdf, 0x63, 0x3d},
		Extension: account.AccountExtension{
			Id:  3,
			Url: "127.0.0.1:8083",
		},
	},
}

func TestNode_Round(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	monkey.Patch(blockchain.InitBlockChain, func(blockchainc.BlockChainConfig, types.EventCenter) error {
		return nil
	})
	var op *port.OutPort
	monkey.PatchInstanceMethod(reflect.TypeOf(op), "BindToPort", func(*port.OutPort, port.OutPutFunc) error {
		return nil
	})
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	assert.NotNil(service)
	node := service.(*Node)

	var r *galaxySolo.SoloPolicy
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "RoleAssignments", func(*galaxySolo.SoloPolicy, []account.Account) (map[account.Account]common.Roler, account.Account, error) {
		return nil, account.Account{}, fmt.Errorf("assignments failed")
	})
	node.Round()

	monkey.PatchInstanceMethod(reflect.TypeOf(r), "RoleAssignments", func(*galaxySolo.SoloPolicy, []account.Account) (map[account.Account]common.Roler, account.Account, error) {
		role := make(map[account.Account]common.Roler)
		role[node.config.Account] = common.Slave
		return role, account.Account{}, nil
	})
	assert.Nil(node.validator)
	node.Round()

	monkey.PatchInstanceMethod(reflect.TypeOf(r), "RoleAssignments", func(*galaxySolo.SoloPolicy, []account.Account) (map[account.Account]common.Roler, account.Account, error) {
		role := make(map[account.Account]common.Roler)
		role[node.config.Account] = common.Master
		return role, account.Account{}, nil
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
	monkey.UnpatchAll()
}

func TestNode_NextRound(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	var op *port.OutPort
	monkey.PatchInstanceMethod(reflect.TypeOf(op), "BindToPort", func(_ *port.OutPort, _ port.OutPutFunc) error {
		return nil
	})
	service, err := NewNode(defaultConf)
	assert.Nil(err)

	bft, err := dbft.NewDBFTPolicy(mockAccounts[0], int64(10))
	assert.Nil(err)
	assert.NotNil(bft)
	node := service.(*Node)
	node.consensus = bft
	monkey.PatchInstanceMethod(reflect.TypeOf(bft), "GetConsensusResult", func(*dbft.DBFTPolicy) gcommon.ConsensusResult {
		return gcommon.ConsensusResult{
			View:        uint64(1),
			Participate: mockAccounts,
			Master:      mockAccounts[1],
		}
	})
	node.NextRound(justitiaCommon.MsgChangeMaster)

	monkey.PatchInstanceMethod(reflect.TypeOf(node), "Round", func(*Node) {
		return
	})
	node.NextRound(justitiaCommon.MsgBlockCommitSuccess)

	bft1, err := fbft.NewFBFTPolicy(mockAccounts[0], int64(10), nil)
	monkey.PatchInstanceMethod(reflect.TypeOf(bft1), "GetConsensusResult", func(*fbft.FBFTPolicy) gcommon.ConsensusResult {
		return gcommon.ConsensusResult{
			View:        uint64(1),
			Participate: mockAccounts,
			Master:      mockAccounts[1],
		}
	})
	node.consensus = bft1
	node.NextRound(justitiaCommon.MsgBlockCommitSuccess)
	monkey.UnpatchAll()
	monkey.UnpatchInstanceMethod(reflect.TypeOf(op), "BindToPort")
}
