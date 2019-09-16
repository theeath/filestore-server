package mq

import "filestore-server/common"

type TransferData struct {
	FileHash string
	CurLocation string
	DestLocation string
	DestStoreType common.StoreType
}