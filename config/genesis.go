package config

import (
	"encoding/json"
	"fmt"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/justitia/tools"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	GenesisFileName = "genesis.json"
	InvalidPath     = ""
)

type GenesisAccountConfig struct {
	Addr    string   `json:"addr"     gencodec:"required"`
	Balance *big.Int `json:"balance"`
	Code    string   `json:"code"`
}

type GenesisBlockConfig struct {
	Block           *types.Block
	GenesisAccounts []GenesisAccountConfig
	ExtraData       []byte `json:"extra_data"`
}

// GenesisAccount is the account in genesis block.
type GenesisAccount struct {
	Addr    types.Address `json:"addr"     gencodec:"required"`
	Balance *big.Int      `json:"balance"`
	Code    []byte        `json:"code"`
}

// GenesisBlock is the genesis block struct of the chain.
type GenesisBlock struct {
	Block           *types.Block
	GenesisAccounts []GenesisAccount
	ExtraData       []byte `json:"extra_data"`
}

// BuildGenesisBlock build genesis block from the specified config file.
// if the genesis config file is not specified, build default genesis block
func GenerateGenesisBlock() (*GenesisBlock, error) {
	var genesisPath = genesisFilePath()
	if InvalidPath == genesisPath {
		log.Info("Start building default genesis block")
		return buildDefaultGenesis()
	} else {
		log.Info("Build genesis block from genesis file: %s", genesisPath)
		return buildGenesisFromConfig(genesisPath)
	}
}

func genesisFilePath() string {
	homePath, _ := tools.Home()
	if tools.PathExists(fmt.Sprintf("%s/.justitia/%s", homePath, GenesisFileName)) {
		return fmt.Sprintf("%s/.justitia/%s", homePath, GenesisFileName)
	}
	goPath := os.Getenv("GOPATH")
	for _, p := range filepath.SplitList(goPath) {
		path := filepath.Join(p, fmt.Sprintf("src/github.com/DSiSc/justitia/config/%s", GenesisFileName))
		if tools.PathExists(path) {
			return path
		}
	}
	return InvalidPath
}

// parse genesis block from config file.
func buildGenesisFromConfig(genesisPath string) (*GenesisBlock, error) {
	file, err := os.Open(genesisPath)
	if err != nil {
		log.Error("Failed to open genesis file, as: %v", err)
		return nil, fmt.Errorf("Failed to open genesis file, as: %v ", err)
	}
	defer file.Close()

	genesis := new(GenesisBlockConfig)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		log.Error("Failed to parse genesis file, as: %v", err)
		return nil, fmt.Errorf("Failed to parse genesis file, as: %v ", err)
	}
	genesisBlock := &GenesisBlock{
		Block:           genesis.Block,
		GenesisAccounts: make([]GenesisAccount, 0),
		ExtraData:       genesis.ExtraData,
	}
	for _, account := range genesis.GenesisAccounts {
		genesisAccount := GenesisAccount{
			Addr:    tools.HexToAddress(account.Addr),
			Balance: account.Balance,
			Code:    tools.Hex2Bytes(account.Code),
		}
		genesisBlock.GenesisAccounts = append(genesisBlock.GenesisAccounts, genesisAccount)
	}
	return genesisBlock, err
}

// build default genesis block.
func buildDefaultGenesis() (*GenesisBlock, error) {
	genesisHeader := &types.Header{
		PrevBlockHash: types.Hash{},
		TxRoot:        types.Hash{},
		ReceiptsRoot:  types.Hash{},
		Height:        uint64(0),
		Timestamp:     uint64(time.Date(2018, time.August, 28, 0, 0, 0, 0, time.UTC).Unix()),
	}

	// genesis block
	genesisBlock := &GenesisBlock{
		Block: &types.Block{
			Header:       genesisHeader,
			Transactions: make([]*types.Transaction, 0),
		},
		ExtraData: nil,
		GenesisAccounts: []GenesisAccount{
			{
				Addr:    tools.HexToAddress("0x0000000000000000000000000000000000000000"),
				Balance: new(big.Int).SetInt64(math.MaxInt64),
			},
			{
				Addr:    tools.HexToAddress("0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
				Balance: new(big.Int).SetInt64(math.MaxInt64),
			},
		},
	}
	return genesisBlock, nil
}