package main

import (
	"flag"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/justitia/common"
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

func argsParse() common.SysConfig {
	logLevel := flag.Int("log_level", common.InvalidInt, "Log level [0: debug, 1: info, 2: warn, 3: error, 4: fatal, 5: panic, 6: disable].")
	logPath := flag.String("log_path", common.BlankString, "Log output file in absolute path.")
	logStyle := flag.String("log_style", common.BlankString, "Log output style in json or text, which choose from [json, text].")
	flag.Parse()
	var style string = *logStyle
	switch style {
	case "text":
		style = log.TextFmt
	case "json":
		style = log.JsonFmt
	}
	return common.SysConfig{
		LogLevel: log.Level(*logLevel),
		LogPath:  *logPath,
		LogStyle: style,
	}
}

func main() {
	node, err := node.NewNode(argsParse())
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
