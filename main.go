package justitia

import (
	"github.com/DSiSc/justitia/node"
	"github.com/DSiSc/txpool/log"
)

func main() {
	node, err := node.NewNode()
	if nil != err {
		log.Error("Failed to instance a node.")
	}
	node.Start()
}
