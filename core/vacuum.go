package core

import (
	"time"
	"bytes"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
)

const vacuumDepth = 65536
var lastVacuumNumber uint64

func vacuum(bc *BlockChain) {
	timer := time.NewTicker(3 * time.Hour)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			vacuumFull(bc)
		case <-bc.quit:
			return
		}
	}
}

func vacuumDiff(bc *BlockChain) {
	currentHeight := bc.hc.CurrentHeader().Number.Uint64()
	if currentHeight < lastVacuumNumber + vacuumDepth + 1 {
		return
	}

	lp := uint64(0)
	n := lastVacuumNumber + 1
	if n == 1 {
		n = currentHeight - 3 * vacuumDepth
	}
	lastVacuumNumber = currentHeight - vacuumDepth - 1
	if lastVacuumNumber < n {
		return
	}

	log.Info("Diff vacuum cleaner", "from", n, "to", lastVacuumNumber)

	for n <= lastVacuumNumber {
		hash := bc.GetCanonicalHash(n)
		if hash != (common.Hash{}) {
			cleanDatabase(bc, n, hash)
		}
		n++
		percent := 100 * n / lastVacuumNumber
		if percent != lp {
			lp = percent
			log.Info("Diff vacuum cleaner", "progress", percent)
		}
	}
}

func vacuumFull(bc *BlockChain) {
	lp := uint64(0)
	log.Info("Full vacuum cleaner", "from", 1, "to", lastVacuumNumber)

	for n := uint64(1); n < lastVacuumNumber; n++ {
		block := bc.GetBlockByNumber(n)
		if block != nil {
			for _, tx := range block.Transactions() {
				rawdb.DeleteTxLookupEntry(bc.db, tx.Hash())
			}
			cleanDatabase(bc, n, block.Root())
		}
		percent := 100 * n / lastVacuumNumber
		if percent != lp {
			lp = percent
			log.Info("Full vacuum cleaner", "progress", percent)
		}
	}
}

func cleanDatabase(bc *BlockChain, number uint64, hash common.Hash) {
	emptyRoot := common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
	emptyCode := crypto.Keccak256(nil)
	batch := bc.db.NewBatch()

	for _, h := range rawdb.ReadAllHashes(bc.db, number) {
		rawdb.DeleteBlock(batch, h, number)
	}

	rawdb.DeleteBlock(batch, hash, number)
	rawdb.DeleteDiffLayer(batch, hash)
	rawdb.WriteHeaderNumber(batch, hash, number)

	trieDB := bc.stateCache.TrieDB()
	if t, err := trie.NewSecure(hash, trieDB); err == nil {
		ait := t.NodeIterator(nil)
		accIter := trie.NewIterator(ait)
		for accIter.Next() {
			var acc state.Account
			if err := rlp.DecodeBytes(accIter.Value, &acc); err == nil {
				if acc.Root != emptyRoot {
					if storageTrie, err := trie.NewSecure(acc.Root, trieDB); err == nil {
						storageIter := storageTrie.NodeIterator(nil)
						for storageIter.Next(true) {
							rawdb.DeleteTrieNode(batch, storageIter.Hash())
						}
					}
				}
				if !bytes.Equal(acc.CodeHash, emptyCode) {
					rawdb.DeleteCode(batch, common.BytesToHash(acc.CodeHash))
				}
			}
		}
		for ait.Next(true) {
			rawdb.DeleteTrieNode(batch, ait.Hash())
		}
	}

	if err := batch.Write(); err != nil {
		log.Error("Failed to delete", "number", number, "hash", hash)
	}
}
