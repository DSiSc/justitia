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
}

func Test_Start(t *testing.T) {
	go service.Start()
	time.Sleep(10 * time.Nanosecond)
	service.Stop()
}
