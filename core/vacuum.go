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

func vacuum(bc *BlockChain) {
	timer := time.NewTicker(3 * time.Minute)
	defer func() {
		bc.wg.Done()
		timer.Stop()
	}()
	for {
		select {
		case <-bc.quit:
			return
		case <-timer.C:
			vacuumCleanDb(bc)
		}
	}
}

func vacuumCleanDb(bc *BlockChain) {
	log.Info("Working vacuum cleaner")
	
	batch := bc.db.NewBatch()
	genesisHash := rawdb.ReadCanonicalHash(bc.db, 0)

	for _, prefix := range rawdb.KeyPrefixSet {
		it := bc.db.NewIterator(prefix, nil)
		for it.Next() {
			if rawdb.IsGenesisKey(genesisHash, it.Key()) {
				continue
			}
			if err := batch.Delete(it.Key()); err != nil {
				log.Warn("Delete from db", "key", it.Key())
			}
		}
		if err := batch.Write(); err != nil {
			log.Error("Failed to clean", "err", err)
		}
		it.Release()
	}
	
	bc.db.PruneAncients()

	log.Info("Finished vacuum cleaner")
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
