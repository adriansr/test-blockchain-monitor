package client

import (
	"sync"

	"monitor/pkg/model"

	"github.com/ethereum/go-ethereum/core/types"
)

type Merger struct {
	mu          sync.Mutex
	lastBlock   uint64
	logsByBlock map[uint64][]types.Log
	blocks      map[uint64]*types.Header
	ready       []model.Block
}

func NewMerger() *Merger {
	return &Merger{
		mu:          sync.Mutex{},
		lastBlock:   0,
		logsByBlock: make(map[uint64][]types.Log),
		blocks:      make(map[uint64]*types.Header),
	}
}

func (r *Merger) AddLog(log types.Log) {
	r.mu.Lock()
	defer r.mu.Unlock()
	blockNum := log.BlockNumber
	if blockNum != r.lastBlock {
		r.deliver(r.lastBlock)
		r.lastBlock = blockNum
	}
	r.logsByBlock[blockNum] = append(r.logsByBlock[blockNum], log)

}

func (r *Merger) deliver(blockNum uint64) {
	if ptr, found := r.blocks[blockNum]; found {
		r.ready = append(r.ready, model.Block{
			Header: *ptr,
			Logs:   r.logsByBlock[blockNum],
		})
		delete(r.blocks, blockNum)
		delete(r.logsByBlock, blockNum)
	}
}

func (r *Merger) AddHead(head *types.Header) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !head.Number.IsUint64() {
		panic(head.Number)
	}
	blockNum := head.Number.Uint64()
	if blockNum != r.lastBlock {
		r.deliver(r.lastBlock)
		r.lastBlock = blockNum
	}
	r.blocks[blockNum] = head
}

func (r *Merger) Blocks() []model.Block {
	r.mu.Lock()
	defer r.mu.Unlock()
	rv := r.ready
	r.ready = nil
	return rv
}
