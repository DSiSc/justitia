package signal

import (
	"github.com/stretchr/testify/assert"
	"os"
	"syscall"
	"testing"
)

func TestNewSignalSet(t *testing.T) {
	assert := assert.New(t)
	ss := NewSignalSet()
	assert.NotNil(ss)
}

var sigint bool = false

func sigintHandler(s os.Signal, arg interface{}) {
	sigint = true
	return
}

func TestSignalSet_RegisterSysSignal(t *testing.T) {
	assert := assert.New(t)
	ss := NewSignalSet()
	ss.RegisterSysSignal(syscall.SIGINT, sigintHandler)
	handler := ss.m[syscall.SIGINT]
	assert.NotNil(handler)
}

/*
func TestSignalSet_CatchSysSignal(t *testing.T) {
	assert := assert.New(t)
	ss := NewSignalSet()
	ss.RegisterSysSignal(syscall.SIGINT, sigintHandler)
	go ss.CatchSysSignal()
	time.Sleep(5*time.Second)
	pid := os.Getegid()
	process, err := os.FindProcess(pid)
	assert.Nil(err)
	process.Signal(syscall.SIGINT)
	assert.True(sigint)
}
*/
