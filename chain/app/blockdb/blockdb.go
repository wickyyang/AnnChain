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

package blockdb

import (
	"fmt"
	"math/big"
	"path/filepath"
	"strings"
	"sync"

	ethcmn "github.com/dappledger/AnnChain/eth/common"
	"github.com/dappledger/AnnChain/eth/common/math"
	"github.com/dappledger/AnnChain/eth/core/rawdb"
	ethstate "github.com/dappledger/AnnChain/eth/core/state"
	ethtypes "github.com/dappledger/AnnChain/eth/core/types"
	"github.com/dappledger/AnnChain/eth/ethdb"
	"github.com/dappledger/AnnChain/eth/rlp"
	"github.com/dappledger/AnnChain/gemmill/modules/go-log"
	"github.com/dappledger/AnnChain/gemmill/modules/go-merkle"
	atypes "github.com/dappledger/AnnChain/gemmill/types"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	OfficialAddress     = "0x7752b42608a0f1943c19fc5802cb027e60b4c911"
	StateRemoveEmptyObj = true
	DatabaseCache       = 128
	DatabaseHandles     = 1024
	APP_NAME            = "blockDB"
)

//reference ethereum BlockChain
type BlockChainDB struct {
	db ethdb.Database
}

type Hashs []ethcmn.Hash
type BeginExecFunc func() (ExecFunc, EndExecFunc)
type ExecFunc func(index int, raw []byte, tx *ethtypes.Transaction) error
type EndExecFunc func(bs []byte, err error) bool

func NewBlockChain(db ethdb.Database) *BlockChainDB {
	return &BlockChainDB{db}
}

func (bc *BlockChainDB) GetHeader(hash ethcmn.Hash, number uint64) *ethtypes.Header {

	//todo cache,reference core/headerchain.go
	header := rawdb.ReadHeader(bc.db, hash, number)
	if header == nil {
		return nil
	}
	return header
}

var (
	ReceiptsPrefix  = []byte("receipt-")
	BlockHashPrefix = []byte("blockhash-")
)

type LastBlockInfo struct {
	Height  int64
	AppHash []byte
}

type BlockDBApp struct {
	atypes.BaseApplication
	AngineHooks atypes.Hooks

	core atypes.Core

	datadir string
	Config  *viper.Viper

	iSyncAllData     bool
	syncDataAccounts map[string]bool

	currentHeader *ethtypes.Header

	stateDb      ethdb.Database
	stateMtx     sync.Mutex
	state        *ethstate.StateDB
	currentState *ethstate.StateDB

	//	valid_hashs Hashs
	sreceipt ethtypes.SReceipts

	Signer ethtypes.Signer
}

var (
	EmptyTrieRoot = ethcmn.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
	lastBlockKey  = []byte("lastblock")
)

func stateKey(block *atypes.Block, height, round int64) string {
	return ethcmn.Bytes2Hex(block.Hash())
}

func OpenDatabase(datadir string, name string, cache int, handles int) (ethdb.Database, error) {
	return ethdb.NewLDBDatabase(filepath.Join(datadir, name), cache, handles)
}

func NewBlockDBApp(config *viper.Viper) (*BlockDBApp, error) {

	syncDataAccounts := make(map[string]bool)
	iSyncAllData := config.GetBool("is_sync_all_data")
	if !iSyncAllData {
		var syncAccounts []string
		configSyncAccountsStr := config.GetString("sync_data_accounts")
		if configSyncAccountsStr != "" {
			syncAccounts = strings.Split(configSyncAccountsStr, ",")
			for _, account := range syncAccounts {
				accountLower := strings.ToLower(account)
				if !strings.HasPrefix(accountLower, "0x") {
					accountLower = fmt.Sprintf("%s%s", "0x", accountLower)
				}
				syncDataAccounts[accountLower] = true
			}
		}
	}

	app := &BlockDBApp{
		datadir:          config.GetString("db_dir"),
		Config:           config,
		iSyncAllData:     iSyncAllData,
		syncDataAccounts: syncDataAccounts,
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

func (app *BlockDBApp) writeGenesis() error {

	if app.getLastAppHash() != EmptyTrieRoot {
		return nil
	}
	app.SaveLastBlock(LastBlockInfo{Height: 0, AppHash: []byte{}})
	return nil
}

func (app *BlockDBApp) Start() (err error) {

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

func (app *BlockDBApp) getLastAppHash() ethcmn.Hash {

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

func (app *BlockDBApp) Stop() {
	app.BaseApplication.Stop()
	app.stateDb.Close()
}

func (app *BlockDBApp) GetAngineHooks() atypes.Hooks {
	return app.AngineHooks
}

func (app *BlockDBApp) CompatibleWithAngine() {}

func (app *BlockDBApp) BeginExecute() {
}

func (app *BlockDBApp) OnNewRound(height, round int64, block *atypes.Block) (interface{}, error) {
	return atypes.NewRoundResult{}, nil
}

func (app *BlockDBApp) OnPrevote(height, round int64, block *atypes.Block) (interface{}, error) {
	return nil, nil
}

func (app *BlockDBApp) OnExecute(height, round int64, block *atypes.Block) (interface{}, error) {
	var (
		res atypes.ExecuteResult
		err error
	)

	exeWithCPUSerialVeirfy(nil, block.Data.Txs, app.genExecFun(block, &res))

	return res, err
}

func makeCurrentHeader(block *atypes.Block, header *atypes.Header) *ethtypes.Header {
	return &ethtypes.Header{
		ParentHash: ethcmn.BytesToHash(block.Header.LastBlockID.Hash),
		Difficulty: big.NewInt(0),
		GasLimit:   math.MaxBig256.Uint64(),
		Time:       big.NewInt(block.Header.Time.Unix()),
		Number:     big.NewInt(header.Height),
	}
}

func (app *BlockDBApp) genExecFun(block *atypes.Block, res *atypes.ExecuteResult) BeginExecFunc {

	app.currentHeader = makeCurrentHeader(block, block.Header)

	return func() (ExecFunc, EndExecFunc) {

		tmpReceipt := make([]*ethtypes.SReceipt, 0)

		execFunc := func(txIndex int, raw []byte, tx *ethtypes.Transaction) error {

			txhash := tx.Hash()

			tmpReceipt = append(tmpReceipt, &ethtypes.SReceipt{Height: app.currentHeader.Number.Uint64(), Timestamp: tx.Timestamp().Uint64(), From: tx.From(), Value: tx.Value(), TxHash: txhash})

			return nil
		}

		endFunc := func(raw []byte, err error) bool {
			if err != nil {
				log.Warn("[evm execute],apply transaction", zap.Error(err))
				tmpReceipt = nil
				res.InvalidTxs = append(res.InvalidTxs, atypes.ExecuteInvalidTx{Bytes: raw, Error: err})
				return true
			}
			app.sreceipt = append(app.sreceipt, tmpReceipt...)
			res.ValidTxs = append(res.ValidTxs, raw)

			return true
		}
		return execFunc, endFunc
	}
}

func exeWithCPUSerialVeirfy(signer ethtypes.Signer, txs atypes.Txs, beginExec BeginExecFunc) error {
	for i, raw := range txs {
		txbs := atypes.Tx(txs[i])
		exec, end := beginExec()
		err := txbs.Deal(func(atx atypes.Tx) error {
			var tx *ethtypes.Transaction
			if len(atx) > 0 {
				tx = new(ethtypes.Transaction)
				if err := rlp.DecodeBytes(atx, tx); err != nil {
					return err
				}
				if err := exec(i, raw, tx); err != nil {
					return err
				}
			}
			return nil
		})
		end(raw, err)
	}

	return nil
}

// OnCommit run in a sync way, we don't need to lock stateDupMtx, but stateMtx is still needed
func (app *BlockDBApp) OnCommit(height, round int64, block *atypes.Block) (interface{}, error) {

	var err error

	app.stateMtx.Lock()
	if app.state, err = ethstate.New(app.getLastAppHash(), ethstate.NewDatabase(app.stateDb)); err != nil {
		app.stateMtx.Unlock()
		return nil, errors.Wrap(err, "create StateDB failed")
	}
	app.stateMtx.Unlock()

	app.SaveLastBlock(LastBlockInfo{Height: height, AppHash: nil})

	rHash := app.SaveValues()

	app.sreceipt = nil

	return atypes.CommitResult{
		AppHash:      nil,
		ReceiptsHash: rHash,
	}, nil
}

func (app *BlockDBApp) CheckTx(bs []byte) error {

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

		if tx.From().Hex() != from.Hex() {
			return errors.New("address and privkey is mismatching")
		}

		return nil
	})
}

func (app *BlockDBApp) SaveValues() []byte {

	savedReceipts := make([][]byte, 0, len(app.sreceipt))

	receiptBatch := app.stateDb.NewBatch()

	for _, receipt := range app.sreceipt {
		storageReceipt := (*ethtypes.SReceiptForStorage)(receipt)

		storageReceiptBytes, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			fmt.Println("wrong rlp encode:" + err.Error())
			return nil
		}

		if app.iSyncAllData {
			if err := receiptBatch.Put(receipt.TxHash.Bytes(), storageReceiptBytes); err != nil {
				fmt.Println("batch receipt failed:" + err.Error())
				return nil
			}
		} else {
			fromAccount := receipt.From.String()
			fromAccountLowerCase := strings.ToLower(fromAccount)
			if app.syncDataAccounts[fromAccountLowerCase] {
				if err := receiptBatch.Put(receipt.TxHash.Bytes(), storageReceiptBytes); err != nil {
					fmt.Println("batch receipt failed:" + err.Error())
					return nil
				}
			}
		}

		savedReceipts = append(savedReceipts, storageReceiptBytes)
	}

	if err := receiptBatch.Write(); err != nil {
		fmt.Println("persist receipts failed:" + err.Error())
		return nil
	}

	rHash := merkle.SimpleHashFromHashes(savedReceipts)

	return rHash
}

func (app *BlockDBApp) Info() (resInfo atypes.ResultInfo) {
	lb := &LastBlockInfo{
		AppHash: make([]byte, 0),
		Height:  0,
	}
	if res, err := app.LoadLastBlock(lb); err == nil {
		lb = res.(*LastBlockInfo)
	}

	resInfo.LastBlockAppHash = lb.AppHash
	resInfo.LastBlockHeight = lb.Height
	resInfo.Version = "0.1.0"
	resInfo.Data = "blockdb"
	return
}

func (app *BlockDBApp) Get(key []byte) atypes.Result {

	data, err := app.stateDb.Get(key)
	if err != nil {
		return atypes.NewError(atypes.CodeType_InternalError, "fail to get receipt for key:"+string(key))
	}
	return atypes.NewResultOK(data, "")
}

func (app *BlockDBApp) SetCore(core atypes.Core) {
	app.core = core
}
