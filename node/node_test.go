package node

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var service NodeService

func Test_NewNode(t *testing.T) {
	var err error
	assert := assert.New(t)
	service, err = NewNode()
	assert.NotNil(service)
	assert.Nil(err)
	go service.Start()
	time.Sleep(time.Duration(10) * time.Microsecond)
	go service.Stop()
}
