package signal

import (
	"fmt"
	"github.com/DSiSc/craft/log"
	"os"
	"os/signal"
)

type signalHandler func(s os.Signal, arg interface{})

type SignalSet struct {
	m map[os.Signal]signalHandler
}

func NewSignalSet() *SignalSet {
	ss := new(SignalSet)
	ss.m = make(map[os.Signal]signalHandler)
	return ss
}

func (set *SignalSet) RegisterSysSignal(s os.Signal, handler signalHandler) {
	if _, found := set.m[s]; !found {
		set.m[s] = handler
	} else {
		log.Error("signal %x has register, please confirm")
	}
}

func (set *SignalSet) handle(sig os.Signal, arg interface{}) (err error) {
	if _, found := set.m[sig]; found {
		set.m[sig](sig, arg)
		return nil
	} else {
		return fmt.Errorf("no handler available for signal %v", sig)
	}
}

func (set *SignalSet) CatchSysSignal() {
	for {
		c := make(chan os.Signal)
		var sigs []os.Signal
		for sig := range set.m {
			sigs = append(sigs, sig)
		}
		signal.Notify(c)
		sig := <-c
		err := set.handle(sig, nil)
		if err != nil {
			fmt.Printf("unknown signal received: %v\n", sig)
			os.Exit(1)
		}
	}
}
