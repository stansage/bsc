package rawdb

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

