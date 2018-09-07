package main

import (
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/justitia/node"
)

func main() {
	node, err := node.NewNode()
	if nil != err {
		log.Error("Failed to initial a node with err %v.", err)
	}
	node.Start()
	node.Wait()
}
