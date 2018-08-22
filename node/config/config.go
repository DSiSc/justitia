package config

import (
	"encoding/json"
	"github.com/DSiSc/producer/config"
	"github.com/DSiSc/txpool/common"
	"github.com/DSiSc/txpool/common/log"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
)

var ConfigName = "config.json"
var DefaultDataDir = "./config"

type Config struct {
	filePath string
	maps     map[string]interface{}
}

type NodeConfig struct {
	// txpool
	GlobalSlots uint64

	// default account
	Account common.Address
}

func NewNodeConfig() NodeConfig {
	// TODO: get account and globalSlots from genesis.json
	var temp common.Address
	var globalSlots uint64 = 10
	return NodeConfig{
		GlobalSlots: globalSlots,
		Account:     temp,
	}
}

func NewProducerConf() config.ProducerConf {
	return config.ProducerConf{
		PolicyName: "timer",
		PolicyContext: config.ProducerPolicy{
			Timer: uint64(10),
			Num:   uint64(0),
		},
	}
}

func New(path string) Config {
	return Config{filePath: path}
}

// Resturn absolute path of config.json
func ConfigAbsPath() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		log.Error("Get config path failed.")
		return file
	}
	keyString := "/github.com/DSiSc/"
	index := strings.LastIndex(file, keyString)
	confAbsPath := strings.Join([]string{file[:index+len(keyString)], "producer/config/config.json"}, "")
	return confAbsPath
}

func DBAbsPath() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		log.Error("Get config path failed.")
		return file
	}
	keyString := "/github.com/DSiSc/producer/"
	index := strings.LastIndex(file, keyString)
	confAbsPath := strings.Join([]string{file[:index+len(keyString)], "config/data"}, "")
	return confAbsPath
}

// Read the given json file.
func (config *Config) read() {
	if !filepath.IsAbs(config.filePath) {
		filePath, err := filepath.Abs(config.filePath)
		if err != nil {
			panic(err)
		}
		config.filePath = filePath
	}

	bts, err := ioutil.ReadFile(config.filePath)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bts, &config.maps)

	if err != nil {
		panic(err)
	}
}

// If we want to get item in a stucture, which like this:
//{
//	"classs": {
//		"student":{
//			"name": "john"
//         }
//     }
//}
// { class: {}}
// You can get it by call Get("class.student.name")
func (config *Config) GetConfigItem(name string) interface{} {
	if config.maps == nil {
		config.read()
	}

	if config.maps == nil {
		return nil
	}

	keys := strings.Split(name, ".")
	length := len(keys)
	if length == 1 {
		return config.maps[name]
	}

	var ret interface{}
	for i := 0; i < length; i++ {
		if i == 0 {
			ret = config.maps[keys[i]]
			if ret == nil {
				return nil
			}
		} else {
			if m, ok := ret.(map[string]interface{}); ok {
				ret = m[keys[i]]
			} else {
				if length == i-1 {
					return ret
				}
				return nil
			}
		}
	}
	return ret
}
