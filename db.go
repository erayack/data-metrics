package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/billy"
)

type DataPool struct {
	id   uint64      
	size uint32      
	store billy.Database
	stored uint64
	nonce uint64
	hash common.Hash
}




