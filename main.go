package main

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/justitia/node"
)

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
	node.Wait()
}
