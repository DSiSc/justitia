package node

import (
	"fmt"
	"github.com/DSiSc/apigateway"
	"github.com/DSiSc/blockchain"
	blockChainConfig "github.com/DSiSc/blockchain/config"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy"
	galaxyCommon "github.com/DSiSc/galaxy/common"
	consensusCommon "github.com/DSiSc/galaxy/consensus/common"
	consensusConfig "github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/consensus/policy/dbft"
	"github.com/DSiSc/galaxy/consensus/policy/fbft"
	"github.com/DSiSc/galaxy/consensus/policy/solo"
	participatesSolo "github.com/DSiSc/galaxy/participates/policy/solo"
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
	p2pConfig "github.com/DSiSc/p2p/config"
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

func TestInitLog(t *testing.T) {
	nodeConfig := config.NodeConfig{
		Logger: log.Config{
			Enabled:         true,
			TimeFieldFormat: "2006-01-02 15:04:05.000",
			Appenders:       make(map[string]*log.Appender),
		},
	}
	nodeConfig.Logger.Appenders[config.FileLogAppender] = &log.Appender{}
	monkey.Patch(log.SetGlobalConfig, func(config *log.Config) {
		return
	})
	InitLog(defaultConf, nodeConfig)
	fileLog := nodeConfig.Logger.Appenders[config.FileLogAppender]
	assert.Equal(t, defaultConf.LogLevel, fileLog.LogLevel)
	assert.Equal(t, defaultConf.LogStyle, fileLog.Format)
	monkey.Unpatch(log.SetGlobalConfig)
}

func TestNewNode(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	monkey.Patch(gossipswitch.NewGossipSwitchByType, func(switchType gossipswitch.SwitchType, _ types.EventCenter) (*gossipswitch.GossipSwitch, error) {
		if gossipswitch.TxSwitch == switchType {
			return nil, fmt.Errorf("mock gossipswitch error")
		}
		return nil, nil
	})
	service, err := NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("txswitch init failed"))
	monkey.Unpatch(gossipswitch.NewGossipSwitchByType)

	var op *port.OutPort
	monkey.PatchInstanceMethod(reflect.TypeOf(op), "BindToPort", func(_ *port.OutPort, _ port.OutPutFunc) error {
		return nil
	})
	monkey.Patch(blockchain.InitBlockChain, func(blockChainConfig.BlockChainConfig, types.EventCenter) error {
		return fmt.Errorf("mock blockchain error")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("blockchain init failed"))

	monkey.Patch(blockchain.InitBlockChain, func(blockChainConfig.BlockChainConfig, types.EventCenter) error {
		return nil
	})
	monkey.Patch(p2p.NewP2P, func(*p2pConfig.P2PConfig, types.EventCenter) (*p2p.P2P, error) {
		return nil, fmt.Errorf("new p2p failed")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("init block syncer p2p failed"))

	monkey.Patch(p2p.NewP2P, func(*p2pConfig.P2PConfig, types.EventCenter) (*p2p.P2P, error) {
		return nil, nil
	})
	monkey.Patch(syncer.NewBlockSyncer, func(p2p.P2PAPI, chan<- interface{}, types.EventCenter) (*syncer.BlockSyncer, error) {
		return nil, fmt.Errorf("new block syncer failed")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("init block syncer failed"))

	monkey.Patch(syncer.NewBlockSyncer, func(p2p.P2PAPI, chan<- interface{}, types.EventCenter) (*syncer.BlockSyncer, error) {
		return nil, nil
	})
	monkey.Patch(propagator.NewBlockPropagator, func(p2p.P2PAPI, chan<- interface{}, types.EventCenter) (*propagator.BlockPropagator, error) {
		return nil, nil
	})
	monkey.Patch(galaxy.NewGalaxyPlugin, func(galaxyCommon.GalaxyPluginConf) (*galaxyCommon.GalaxyPlugin, error) {
		return nil, fmt.Errorf("error of NewGalaxyPlugin")
	})
	service, err = NewNode(defaultConf)
	assert.Equal(err, fmt.Errorf("init galaxy plugin failed with error error of NewGalaxyPlugin"))
	assert.NotNil(service)

	monkey.Patch(galaxy.NewGalaxyPlugin, func(galaxyCommon.GalaxyPluginConf) (*galaxyCommon.GalaxyPlugin, error) {
		return nil, nil
	})
	nodeConf := config.NewNodeConfig()
	monkey.Patch(config.NewNodeConfig, func() config.NodeConfig {
		nodeConf.NodeType = justitiaCommon.FullNode
		return nodeConf
	})
	service, err = NewNode(defaultConf)
	nodeService := service.(*Node)
	event := nodeService.eventCenter.(*events.Event)
	assert.Equal(1, len(event.Subscribers))
	assert.NotNil(service)
	monkey.Unpatch(blockchain.InitBlockChain)
	monkey.UnpatchInstanceMethod(reflect.TypeOf(op), "BindToPort")
	monkey.Unpatch(p2p.NewP2P)
	monkey.Unpatch(syncer.NewBlockSyncer)
	monkey.Unpatch(propagator.NewBlockPropagator)
	monkey.Unpatch(galaxy.NewGalaxyPlugin)
	monkey.Unpatch(InitLog)
	monkey.Unpatch(config.NewNodeConfig)
	monkey.Unpatch(config.GetLogSetting)
}

func TestNode_Start(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(blockchain.InitBlockChain, func(blockChainConfig.BlockChainConfig, types.EventCenter) error {
		return nil
	})
	monkey.Patch(syncer.NewBlockSyncer, func(p2p.P2PAPI, chan<- interface{}, types.EventCenter) (*syncer.BlockSyncer, error) {
		return nil, nil
	})
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	assert.NotNil(service)
	var ch = make(chan int)
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
	go func() {
		service.Start()
		nodeService := service.(*Node)
		assert.NotNil(nodeService.rpcListeners)
		assert.Equal(0, len(nodeService.rpcListeners))
		ch <- 1
	}()
	<-ch
	service.Stop()
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
	monkey.Patch(syncer.NewBlockSyncer, func(p2p.P2PAPI, chan<- interface{}, types.EventCenter) (*syncer.BlockSyncer, error) {
		return nil, nil
	})
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	monkey.Patch(blockchain.InitBlockChain, func(blockChainConfig.BlockChainConfig, types.EventCenter) error {
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
	monkey.PatchInstanceMethod(reflect.TypeOf(c), "ToConsensus", func(*solo.SoloPolicy, *consensusCommon.Proposal) error {
		return fmt.Errorf("consensus failed")
	})
	node.Round()
	monkey.UnpatchAll()
}

func TestNode_NextRound(t *testing.T) {
	var timeout = consensusConfig.ConsensusTimeout{
		TimeoutToChangeView: int64(1000),
	}
	assert := assert.New(t)
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	service, err := NewNode(defaultConf)
	assert.Nil(err)

	bft, err := dbft.NewDBFTPolicy(mockAccounts[0], timeout)
	assert.Nil(err)
	assert.NotNil(bft)
	node := service.(*Node)
	node.consensus = bft
	monkey.PatchInstanceMethod(reflect.TypeOf(bft), "GetConsensusResult", func(*dbft.DBFTPolicy) consensusCommon.ConsensusResult {
		return consensusCommon.ConsensusResult{
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

	bft1, err := fbft.NewFBFTPolicy(mockAccounts[0], timeout, nil, true, consensusConfig.SignatureVerifySwitch{})
	monkey.PatchInstanceMethod(reflect.TypeOf(bft1), "GetConsensusResult", func(*fbft.FBFTPolicy) consensusCommon.ConsensusResult {
		return consensusCommon.ConsensusResult{
			View:        uint64(1),
			Participate: mockAccounts,
			Master:      mockAccounts[1],
		}
	})
	node.consensus = bft1
	node.NextRound(justitiaCommon.MsgBlockCommitSuccess)
	monkey.UnpatchAll()
}

func TestNode_OnlineWizard(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	node := service.(*Node)
	var p *participatesSolo.SoloPolicy
	monkey.PatchInstanceMethod(reflect.TypeOf(p), "GetParticipates", func(policy *participatesSolo.SoloPolicy) ([]account.Account, error) {
		return make([]account.Account, 0), fmt.Errorf("get participte failed")
	})
	node.OnlineWizard()

	monkey.PatchInstanceMethod(reflect.TypeOf(p), "GetParticipates", func(policy *participatesSolo.SoloPolicy) ([]account.Account, error) {
		return make([]account.Account, 0), nil
	})
	var r *galaxySolo.SoloPolicy
	monkey.PatchInstanceMethod(reflect.TypeOf(r), "RoleAssignments", func(*galaxySolo.SoloPolicy, []account.Account) (map[account.Account]common.Roler, account.Account, error) {
		return nil, account.Account{}, fmt.Errorf("assignments failed")
	})
	node.OnlineWizard()

	monkey.PatchInstanceMethod(reflect.TypeOf(r), "RoleAssignments", func(*galaxySolo.SoloPolicy, []account.Account) (map[account.Account]common.Roler, account.Account, error) {
		return nil, account.Account{}, nil
	})
	var c *solo.SoloPolicy
	monkey.PatchInstanceMethod(reflect.TypeOf(c), "Initialization", func(*solo.SoloPolicy, account.Account, []account.Account, types.EventCenter, bool) {
		return
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(c), "Online", func(*solo.SoloPolicy) {
		return
	})
	node.OnlineWizard()
	monkey.UnpatchAll()
}

func TestNode_Wait(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	node := service.(*Node)
	go func() {
		node.serviceChannel <- uint8(1)
	}()
	node.Wait()
	monkey.UnpatchAll()
}

func TestNewNode2(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	node := service.(*Node)
	node.eventsRegister()
	go func() {
		err := node.eventCenter.Notify(types.EventBlockCommitted, nil)
		assert.Nil(err)
	}()
	ch := <-node.msgChannel
	assert.Equal(justitiaCommon.MsgBlockCommitSuccess, ch)
	monkey.UnpatchAll()
}

func MockNewTrans() []*types.Transaction {
	txs := make([]*types.Transaction, 0, 11)
	var tx *types.Transaction
	for i := 0; i < 11; i++ {
		tx = &types.Transaction{
			Data: types.TxData{
				AccountNonce: uint64(i),
			},
		}
		txs = append(txs, tx)
	}
	return txs
}

func TestNewNode3(t *testing.T) {
	assert := assert.New(t)
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(config.GetLogSetting, func(*viper.Viper) log.Config {
		return log.Config{}
	})
	monkey.Patch(InitLog, func(justitiaCommon.SysConfig, config.NodeConfig) {
		return
	})
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	node := service.(*Node)
	node.eventsRegister()
	block := &types.Block{
		Header: &types.Header{
			Height: uint64(1),
		},
		Transactions: make([]*types.Transaction, 0),
	}
	node.config.NodeType = justitiaCommon.FullNode
	err = node.eventCenter.Notify(types.EventBlockCommitted, block)
	assert.Nil(err)
	monkey.UnpatchAll()
}
