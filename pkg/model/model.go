package model

import (
	"github.com/ethereum/go-ethereum/core/types"
)

type Block struct {
	Header types.Header `json:"header"`
	Logs   []types.Log  `json:"logs"`
}

type Observer[T any] interface {
	OnEvent(t T) error
}
