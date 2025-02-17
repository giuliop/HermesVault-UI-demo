package models

import (
	"github.com/algorand/go-algorand-sdk/v2/abi"
	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/giuliop/algoplonk"
)

type TreeConfig struct {
	Depth      int
	ZeroValue  []byte
	ZeroHashes [][]byte
	HashFunc   func(...[]byte) []byte
}

type App struct {
	Id                 uint64
	Schema             *Arc32Schema
	TSS                *Lsig
	DepositCc          *algoplonk.CompiledCircuit
	WithdrawalCc       *algoplonk.CompiledCircuit
	DepositVerifier    *Lsig
	WithdrawalVerifier *Lsig
	TreeConfig         TreeConfig
}

type Lsig struct {
	Account crypto.LogicSigAccount
	Address types.Address
}

// Arc32Schema defines a partial ARC32 schema
type Arc32Schema struct {
	Source struct {
		Approval string `json:"approval"`
		Clear    string `json:"clear"`
	} `json:"source"`
	State struct {
		Global struct {
			NumByteSlices uint64 `json:"num_byte_slices"`
			NumUints      uint64 `json:"num_uints"`
		} `json:"global"`
		Local struct {
			NumByteSlices uint64 `json:"num_byte_slices"`
			NumUints      uint64 `json:"num_uints"`
		} `json:"local"`
	} `json:"state"`
	Contract abi.Contract `json:"contract"`
}
