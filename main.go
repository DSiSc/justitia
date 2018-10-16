package main

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/justitia/node"
	"github.com/DSiSc/justitia/tools/signal"
	"os"
	"syscall"
)

func sysSignalProcess(node node.NodeService) {
	sysSignalProcess := signal.NewSignalSet()
	sysSignalProcess.RegisterSysSignal(syscall.SIGINT, func(os.Signal, interface{}) {
		log.Warn("handle signal SIGINT.")
		node.Stop()
		os.Exit(1)
	})
	sysSignalProcess.RegisterSysSignal(syscall.SIGTERM, func(os.Signal, interface{}) {
		log.Warn("handle signal SIGTERM.")
		node.Stop()
		os.Exit(1)
	})
	go sysSignalProcess.CatchSysSignal()
}

func main() {
	node, err := node.NewNode()
	if nil != err {
		log.Fatal("Failed to initial a node with err %v.", err)
	}
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("Fatal error occur: %v.", err)
		}
	}()
	node.Start()
	sysSignalProcess(node)
	node.Wait()
}
