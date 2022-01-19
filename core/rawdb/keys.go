package rawdb

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
)

var KeyPrefixSet = map[string][]byte{
	string(headerPrefix):          headerPrefix,
	string(headerNumberPrefix):    headerNumberPrefix,
	string(blockBodyPrefix):       blockBodyPrefix,
	string(blockReceiptsPrefix):   blockReceiptsPrefix,
	string(txLookupPrefix):        txLookupPrefix,
	string(bloomBitsPrefix):       bloomBitsPrefix,
	string(SnapshotAccountPrefix): SnapshotAccountPrefix,
	string(SnapshotStoragePrefix): SnapshotStoragePrefix,
	string(CodePrefix):            CodePrefix,
	string(diffLayerPrefix):       diffLayerPrefix,
	string(preimagePrefix):        preimagePrefix,
	string(configPrefix):          configPrefix,
	string(BloomBitsIndexPrefix):  BloomBitsIndexPrefix,
}


func IsGenesisKey(genesisHash common.Hash, key []byte) bool {
	return bytes.Equal(key, headerKey(0, genesisHash)) ||
		bytes.Equal(key, headerTDKey(0, genesisHash)) ||
		bytes.Equal(key, headerHashKey(0)) ||
		bytes.Equal(key, headerNumberKey(genesisHash)) ||
		bytes.Equal(key, blockBodyKey(0, genesisHash)) ||
		bytes.Equal(key, blockReceiptsKey(0, genesisHash)) ||
		bytes.Equal(key, diffLayerKey(genesisHash)) ||
		bytes.Equal(key, txLookupKey(genesisHash)) ||
		bytes.Equal(key, accountSnapshotKey(genesisHash)) ||
		bytes.Equal(key, preimageKey(genesisHash)) ||
		bytes.Equal(key, codeKey(genesisHash)) ||
		bytes.Equal(key, configKey(genesisHash))
}
