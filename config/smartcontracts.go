package config

import (
	"github.com/consensys/gnark-crypto/ecc"
)

// app constants
const (
	MerkleTreeLevels    = 24
	Curve               = ecc.BN254
	RandomNonceByteSize = 31

	DepositMinimumAmount = 1e6  // 1 algo
	WithDrawalFeeDivisor = 1000 // 0.1% (we divide by this to get the fee)
	WithdrawalMinimumFee = 1e5  // 0.1 algo

	DepositMethodName    = "deposit"
	WithDrawalMethodName = "withdraw"
	NoOpMethodName       = "noop"

	UserDepositTxnIndex = 1 // index of the user pay txn in the deposit txn group (0 based)
)

// transaction fees required
const (
	// # top level transactions needed for logicsig verifier opcode budget
	VerifierTopLevelTxnNeeded = 8

	// fees needed for a deposit transaction group
	DepositMinFeeMultiplier = 42

	// fees needed for a withdrawal transaction group
	WithdrawalMinFeeMultiplier = 47
)

var Hash = NewMimcF(Curve)
