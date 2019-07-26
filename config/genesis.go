package config

import (
	"encoding/json"
	"fmt"
	types2 "github.com/DSiSc/apigateway/core/types"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	justitiac "github.com/DSiSc/justitia/common"
	"github.com/DSiSc/justitia/compiler"
	"github.com/DSiSc/justitia/tools"
	"github.com/DSiSc/repository"
	"github.com/DSiSc/validator/worker"
	"github.com/DSiSc/validator/worker/common"
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
	Addr     string   `json:"addr"     gencodec:"required"`
	Balance  *big.Int `json:"balance"`
	Code     string   `json:"code"`
	Contract string   `json:"contract"`
}

type GenesisBlockConfig struct {
	Block           *types.Block
	GenesisAccounts []GenesisAccountConfig
	ExtraData       []byte `json:"extra_data"`
}

// GenesisAccount is the account in genesis block.
type GenesisAccount struct {
	Addr     types.Address `json:"addr"     gencodec:"required"`
	Balance  *big.Int      `json:"balance"`
	Code     []byte        `json:"code"`
	Contract string        `json:"contract"`
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

// add tx to genesis block
func (genesis *GenesisBlock) addTxToGenesisBlock() {
	var nonce uint64
	for _, key := range genesis.GenesisAccounts {
		if 0 != len(key.Code) {
			tx := types2.NewTransaction(nonce, nil, big.NewInt(0), uint64(0), big.NewInt(0), key.Code, types2.Address{})
			genesis.Block.Transactions = append(genesis.Block.Transactions, tx)
			nonce++
		}
	}
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
		contractByteCode := ""
		genesisAccount := GenesisAccount{
			Addr:     tools.HexToAddress(account.Addr),
			Balance:  account.Balance,
			Contract: account.Contract,
		}
		if contractByteCode != account.Code {
			genesisAccount.Code = tools.Hex2Bytes(account.Code)
		} else {
			if contractByteCode != account.Contract {
				contractByteCode = compiler.SolidityCompile(account.Contract)
				genesisAccount.Code = tools.Hex2Bytes(contractByteCode)
			}
		}
		genesisBlock.GenesisAccounts = append(genesisBlock.GenesisAccounts, genesisAccount)
	}
	genesisBlock.addTxToGenesisBlock()
	genesisBlock.Block.Header.Timestamp = uint64(time.Date(2018, time.August, 28, 0, 0, 0, 0, time.UTC).Unix())
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

func ImportGenesisBlock() {
	chain, err := repository.NewLatestStateRepository()
	if err != nil {
		panic(fmt.Errorf("failed to create init-state block chain, as: %v", err))
	}

	currentBlock := chain.GetCurrentBlock()
	if currentBlock != nil {
		log.Info("found latest block with height %d from local database, will skip importing genesis block", currentBlock.Header.Height)
		return
	}

	genesisBlock, err := GenerateGenesisBlock()
	if err != nil {
		panic(fmt.Errorf("get genesis block failed with error %s", err))
	}

	// set balance and record map
	for _, account := range genesisBlock.GenesisAccounts {
		if nil != account.Balance && account.Balance.Cmp(big.NewInt(0)) == 1 {
			chain.CreateAccount(account.Addr)
			chain.SetBalance(account.Addr, account.Balance)
		}
		if len(account.Code) != 0 {
			contractType := justitiac.SystemContractType(account.Contract)
			if types.InitialContractType == contractType {
				panic("illegal parameter")
			}
		}
	}
	// execute transaction
	for _, tx := range genesisBlock.Block.Transactions {
		_, _, _, err, _ := worker.ApplyTransaction(genesisBlock.Block.Header.Coinbase, genesisBlock.Block.Header, chain, tx, new(common.GasPool))
		if err != nil {
			panic("apply transaction failed")
		}
	}
	// update block header hash
	genesisBlock.Block.HeaderHash = justitiac.HeaderHash(genesisBlock.Block)
	genesisBlock.Block.Header.StateRoot = chain.IntermediateRoot(false)
	// write block
	err = chain.WriteBlock(genesisBlock.Block)
	if nil != err {
		panic("import genesis block failed.")
	}
}

func GetChainIdFromConfig() (uint64, error) {
	var genesisPath = genesisFilePath()
	if InvalidPath == genesisPath {
		log.Info("GenesisPath is invalid, return the default chainId")
		return 0, nil
	}

	//Open genesisFile by the path
	file, err := os.Open(genesisPath)
	if err != nil {
		log.Error("Failed to open genesis file, as: %v", err)
		return 0, fmt.Errorf("Failed to open genesis file, as: %v ", err)
	}
	defer file.Close()

	//Decode the file to genesis
	genesis := new(GenesisBlockConfig)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		log.Error("Failed to parse genesis file, as: %v", err)
		return 0, fmt.Errorf("Failed to parse genesis file, as: %v ", err)
	}
	chainId := genesis.Block.Header.ChainID

	return chainId, nil
}
