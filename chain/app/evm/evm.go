// Copyright 2017 ZhongAn Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package evm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"path/filepath"
	"sync"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	ethcmn "github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/common/math"
	ethcore "github.com/dappledger/AnnChain/eth/core"
	"github.com/dappledger/AnnChain/eth/core/rawdb"
	ethstate "github.com/dappledger/AnnChain/eth/core/state"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/ethdb"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/modules/go-merkle"
	atypes "github.com/dappledger/AnnChain/gemmill/types"
)

const (
	OfficialAddress     = "0x7752b42608a0f1943c19fc5802cb027e60b4c911"
	StateRemoveEmptyObj = true
	DatabaseCache       = 128
	DatabaseHandles     = 1024
	APP_NAME            = "evm"
)

//reference ethereum BlockChain
type BlockChainEvm struct {
	db ethdb.Database
}

type Hashs []ethcmn.Hash
type BeginExecFunc func() (ExecFunc, EndExecFunc)
type ExecFunc func(index int, raw []byte, tx *ethtypes.Transaction) error
type EndExecFunc func(bs []byte, err error) bool

func NewBlockChain(db ethdb.Database) *BlockChainEvm {
	return &BlockChainEvm{db}
}
func (bc *BlockChainEvm) GetHeader(hash ethcmn.Hash, number uint64) *ethtypes.Header {

	//todo cache,reference core/headerchain.go
	header := rawdb.ReadHeader(bc.db, hash, number)
	if header == nil {
		return nil
	}
	return header
}

var (
	ReceiptsPrefix  = []byte("receipts-")
	BlockHashPrefix = []byte("blockhash-")
)

type LastBlockInfo struct {
	Height  int64
	AppHash []byte
}

type EVMApp struct {
	atypes.BaseApplication
	AngineHooks atypes.Hooks

	core atypes.Core

	datadir string
	Config  *viper.Viper

	currentHeader *ethtypes.Header
	//	chainConfig   *ethparams.ChainConfig

	stateDb      ethdb.Database
	stateMtx     sync.Mutex
	state        *ethstate.StateDB
	currentState *ethstate.StateDB

	valid_hashs Hashs

	Signer ethtypes.Signer
}

const (
	// With 2.2 GHz Intel Core i7, 16 GB 2400 MHz DDR4, 256GB SSD, we tested following contract, it takes about 24157 gas and 171.193µs.
	// function setVal(uint256 _val) public {
	//	val = _val;
	//	emit SetVal(_val,_val);
	//  emit SetValByWho("a name which length is bigger than 32 bytes",msg.sender, _val);
	// }
	// So we estimate that running out of 100000000 gas may be taken at least 1s to 10s
	EVMGasLimit uint64 = 100000000
)

var (
	EmptyTrieRoot = ethcmn.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	lastBlockKey = []byte("lastblock")

	errQuitExecute = fmt.Errorf("quit executing block")
)

func makeCurrentHeader(block *atypes.Block, header *atypes.Header) *ethtypes.Header {
	return &ethtypes.Header{
		ParentHash: ethcmn.BytesToHash(block.Header.LastBlockID.Hash),
		Difficulty: big.NewInt(0),
		GasLimit:   math.MaxBig256.Uint64(),
		Time:       big.NewInt(block.Header.Time.Unix()),
		Number:     big.NewInt(header.Height),
	}
}

func stateKey(block *atypes.Block, height, round int64) string {
	return ethcmn.Bytes2Hex(block.Hash())
}

func OpenDatabase(datadir string, name string, cache int, handles int) (ethdb.Database, error) {
	return ethdb.NewLDBDatabase(filepath.Join(datadir, name), cache, handles)
}

func NewEVMApp(config *viper.Viper) (*EVMApp, error) {
	app := &EVMApp{
		datadir: config.GetString("db_dir"),
		Config:  config,
	}

	app.AngineHooks = atypes.Hooks{
		OnNewRound: atypes.NewHook(app.OnNewRound),
		OnCommit:   atypes.NewHook(app.OnCommit),
		OnPrevote:  atypes.NewHook(app.OnPrevote),
		OnExecute:  atypes.NewHook(app.OnExecute),
	}

	app.Signer = new(ethtypes.HomesteadSigner)
	var err error
	if err = app.BaseApplication.InitBaseApplication(APP_NAME, app.datadir); err != nil {
		log.Error("InitBaseApplication error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}

	if app.stateDb, err = OpenDatabase(app.datadir, "chaindata", DatabaseCache, DatabaseHandles); err != nil {
		log.Error("OpenDatabase error", zap.Error(err))
		return nil, errors.Wrap(err, "app error")
	}

	return app, nil
}

func (app *EVMApp) writeGenesis() error {

	//	switch ethcore.GetEvmLimitType() {
	//	case ethcore.EvmLimitTypeTx, ethcore.EvmLimitTypeBalance:
	//	default:
	//		return nil
	//	}

	if app.getLastAppHash() != EmptyTrieRoot {
		return nil
	}

	g := ethcore.DefaultGenesis()
	b := g.ToBlock(app.stateDb)
	app.SaveLastBlock(LastBlockInfo{Height: 0, AppHash: b.Root().Bytes()})
	return nil
}

func (app *EVMApp) Start() (err error) {

	if err := app.writeGenesis(); err != nil {
		app.Stop()
		log.Error("write genesis err:", zap.Error(err))
		return err
	}

	lastBlock := &LastBlockInfo{
		Height:  0,
		AppHash: make([]byte, 0),
	}
	if res, err := app.LoadLastBlock(lastBlock); err == nil && res != nil {
		lastBlock = res.(*LastBlockInfo)
	}
	if err != nil {
		log.Error("fail to load last block", zap.Error(err))
		return
	}

	// Load evm state when starting
	trieRoot := EmptyTrieRoot
	if len(lastBlock.AppHash) > 0 {
		trieRoot = ethcmn.BytesToHash(lastBlock.AppHash)
	}

	if app.state, err = ethstate.New(trieRoot, ethstate.NewDatabase(app.stateDb)); err != nil {
		app.Stop()
		log.Error("fail to new ethstate", zap.Error(err))
		return
	}

	return nil
}

func (app *EVMApp) getLastAppHash() ethcmn.Hash {
	lastBlock := &LastBlockInfo{
		Height:  0,
		AppHash: make([]byte, 0),
	}
	if res, err := app.LoadLastBlock(lastBlock); err == nil && res != nil {
		lastBlock = res.(*LastBlockInfo)
	}
	if len(lastBlock.AppHash) > 0 {
		return ethcmn.BytesToHash(lastBlock.AppHash)
	}
	return EmptyTrieRoot
}

func (app *EVMApp) Stop() {
	app.BaseApplication.Stop()
	app.stateDb.Close()
}

func (app *EVMApp) GetAngineHooks() atypes.Hooks {
	return app.AngineHooks
}

func (app *EVMApp) CompatibleWithAngine() {}

func (app *EVMApp) BeginExecute() {
}

func (app *EVMApp) OnNewRound(height, round int64, block *atypes.Block) (interface{}, error) {
	return atypes.NewRoundResult{}, nil
}

func (app *EVMApp) OnPrevote(height, round int64, block *atypes.Block) (interface{}, error) {
	return nil, nil
}

func (app *EVMApp) OnExecute(height, round int64, block *atypes.Block) (interface{}, error) {
	var (
		res atypes.ExecuteResult
		err error
	)

	if app.currentState, err = ethstate.New(app.getLastAppHash(), ethstate.NewDatabase(app.stateDb)); err != nil {
		return nil, errors.Wrap(err, "create StateDB failed")
	}
	//	exeWithCPUSerialVeirfy(nil, block.Data.Txs, app.genExecFun(block, &res))

	return res, err
}

// OnCommit run in a sync way, we don't need to lock stateDupMtx, but stateMtx is still needed
func (app *EVMApp) OnCommit(height, round int64, block *atypes.Block) (interface{}, error) {
	appHash, err := app.currentState.Commit(StateRemoveEmptyObj)
	if err != nil {
		return nil, err
	}

	if err := app.currentState.Database().TrieDB().Commit(appHash, false); err != nil {
		return nil, err
	}

	app.stateMtx.Lock()
	if app.state, err = ethstate.New(appHash, ethstate.NewDatabase(app.stateDb)); err != nil {
		app.stateMtx.Unlock()
		return nil, errors.Wrap(err, "create StateDB failed")
	}
	app.stateMtx.Unlock()

	app.SaveLastBlock(LastBlockInfo{Height: height, AppHash: appHash.Bytes()})
	//	rHash := app.SaveReceipts()
	bHash := app.SaveBlocks(block.Hash())
	app.valid_hashs = nil

	//	log.Info("application save to db", zap.String("appHash", fmt.Sprintf("%X", appHash.Bytes())), zap.String("receiptHash", fmt.Sprintf("%X", rHash)))

	return atypes.CommitResult{
		AppHash: appHash.Bytes(),
		//		ReceiptsHash: rHash,
		BlockHash: bHash,
	}, nil
}

func (app *EVMApp) CheckTx(bs []byte) error {
	return atypes.Tx(bs).Deal(func(txbs atypes.Tx) error {
		tx := &ethtypes.Transaction{}
		err := rlp.DecodeBytes(txbs, tx)
		if err != nil {
			return err
		}
		from, err := ethtypes.Sender(app.Signer, tx)
		if err != nil {
			return err
		}

		app.stateMtx.Lock()
		defer app.stateMtx.Unlock()
		// Last but not least check for nonce errors
		nonce := tx.Nonce()
		getNonce := app.state.GetNonce(from)
		if getNonce > nonce {
			txhash := atypes.Tx(bs).Hash()
			return fmt.Errorf("nonce(%d) different with getNonce(%d), transaction already exists %v", nonce, getNonce, hex.EncodeToString(txhash))
		}
		// Transactor should have enough funds to cover the costs
		// cost == V + GP * GL
		if app.state.GetBalance(from).Cmp(tx.Cost()) < 0 {
			return fmt.Errorf("not enough funds")
		}
		return nil
	})
}

func (app *EVMApp) SaveBlocks(blockHash []byte) []byte {

	blockBatch := app.stateDb.NewBatch()

	storageBlockBytes, err := rlp.EncodeToBytes(app.valid_hashs)
	if err != nil {
		fmt.Println("wrong rlp encode:" + err.Error())
		return nil
	}

	key := append(BlockHashPrefix, blockHash...)

	if err := blockBatch.Put(key, storageBlockBytes); err != nil {
		fmt.Println("batch block failed:" + err.Error())
		return nil
	}

	if err := blockBatch.Write(); err != nil {
		fmt.Println("persist block failed:" + err.Error())
		return nil
	}

	bHash := merkle.SimpleHashFromBinaries([]interface{}{app.valid_hashs})

	return bHash
}

//func (app *EVMApp) SaveReceipts() []byte {
//	savedReceipts := make([][]byte, 0, len(app.receipts))
//	receiptBatch := app.stateDb.NewBatch()

//	for _, receipt := range app.receipts {
//		storageReceipt := (*ethtypes.ReceiptForStorage)(receipt)
//		savedReceipts = append(savedReceipts, storageReceiptBytes)
//	}
//	if err := receiptBatch.Write(); err != nil {
//		fmt.Println("persist receipts failed:" + err.Error())
//		return nil
//	}

//	rHash := merkle.SimpleHashFromHashes(savedReceipts)

//	return rHash
//}

func (app *EVMApp) Info() (resInfo atypes.ResultInfo) {
	lb := &LastBlockInfo{
		AppHash: make([]byte, 0),
		Height:  0,
	}
	if res, err := app.LoadLastBlock(lb); err == nil {
		lb = res.(*LastBlockInfo)
	}

	resInfo.LastBlockAppHash = lb.AppHash
	resInfo.LastBlockHeight = lb.Height
	resInfo.Version = "0.7.0"
	resInfo.Data = "default app with evm-1.8.21"
	return
}

func (app *EVMApp) Query(query []byte) atypes.Result {
	var res atypes.Result
	// .....
	// .....
	return res
}

func (app *EVMApp) SetCore(core atypes.Core) {
	app.core = core
}
