package node

import (
	"fmt"
	"github.com/DSiSc/apigateway"
	"github.com/DSiSc/blockchain"
	blockchainc "github.com/DSiSc/blockchain/config"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/galaxy/consensus"
	consensusc "github.com/DSiSc/galaxy/consensus/config"
	"github.com/DSiSc/galaxy/participates"
	"github.com/DSiSc/galaxy/participates/config"
	"github.com/DSiSc/galaxy/role"
	rolec "github.com/DSiSc/galaxy/role/config"
	"github.com/DSiSc/gossipswitch"
	commonc "github.com/DSiSc/justitia/common"
	"github.com/DSiSc/justitia/tools/events"
	"github.com/DSiSc/monkey"
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
	/*
		var oport *gossipswitch.OutPort
		monkey.PatchInstanceMethod(reflect.TypeOf(oport), "BindToPort", func (_ *gossipswitch.OutPort, _ gossipswitch.OutPutFunc) error {
			return fmt.Errorf("bind error")
		})
		service, err = NewNode(defaultConf)
		assert.NotNil(err)
		assert.Nil(service)
		assert.Equal(err, fmt.Errorf("registe txpool failed"))
		monkey.UnpatchInstanceMethod(reflect.TypeOf(oport), "BindToPort")
	*/
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

	monkey.Patch(role.NewRole, func(_ participates.Participates, _ account.Account, _ rolec.RoleConfig) (role.Role, error) {
		return nil, fmt.Errorf("mock role error")
	})
	service, err = NewNode(defaultConf)
	assert.NotNil(err)
	assert.Nil(service)
	assert.Equal(err, fmt.Errorf("role init failed"))
	monkey.Unpatch(role.NewRole)

	monkey.Patch(consensus.NewConsensus, func(_ participates.Participates, _ consensusc.ConsensusConfig) (consensus.Consensus, error) {
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
	monkey.Patch(apigateway.StartRPC, func(string) ([]net.Listener, error) {
		return make([]net.Listener, 0), nil
	})
	service, err := NewNode(defaultConf)
	assert.Nil(err)
	assert.NotNil(service)
	go func() {
		service.Start()
		nodeService := service.(*Node)
		assert.NotNil(nodeService.rpcListeners)
		assert.Equal(0, len(nodeService.rpcListeners))
		service.Wait()
	}()
	service.Stop()
}

func TestEventRegister(t *testing.T) {
	types.GlobalEventCenter = events.NewEvent()
	EventRegister()
	event := types.GlobalEventCenter.(*events.Event)
	assert.Equal(t, 3, len(event.Subscribers))
}

func TestEventUnregister(t *testing.T) {
	assert := assert.New(t)
	// init event center
	types.GlobalEventCenter = events.NewEvent()
	EventRegister()
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
