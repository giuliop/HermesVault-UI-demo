package circuits

import (
	"webapp/config"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/accumulator/merkle"
	"github.com/consensys/gnark/std/hash/mimc"
)

const MerkleTreeLevels = config.MerkleTreeLevels

type WithdrawalCircuit struct {
	Recipient  frontend.Variable `gnark:",public"`
	Withdrawal frontend.Variable `gnark:",public"`
	Fee        frontend.Variable `gnark:",public"`
	Commitment frontend.Variable `gnark:",public"`
	Nullifier  frontend.Variable `gnark:",public"`
	Root       frontend.Variable `gnark:",public"`
	K          frontend.Variable
	R          frontend.Variable
	Amount     frontend.Variable
	Change     frontend.Variable
	K2         frontend.Variable
	R2         frontend.Variable
	Index      frontend.Variable
	Path       [MerkleTreeLevels + 1]frontend.Variable
}

func (c *WithdrawalCircuit) Define(api frontend.API) error {
	mimc, _ := mimc.NewMiMC(api)

	// hash(Amount,K) == Nullifier
	mimc.Write(c.Amount)
	mimc.Write(c.K)
	api.AssertIsEqual(c.Nullifier, mimc.Sum())

	mimc.Reset()

	// hash(hash(Change, K2, R2)) == Commitment
	mimc.Write(c.Change)
	mimc.Write(c.K2)
	mimc.Write(c.R2)
	h := mimc.Sum()

	mimc.Reset()

	mimc.Write(h)
	api.AssertIsEqual(c.Commitment, mimc.Sum())

	mimc.Reset()

	// Path[0] == hash(Amount, K, R)
	mimc.Write(c.Amount)
	mimc.Write(c.K)
	mimc.Write(c.R)
	api.AssertIsEqual(c.Path[0], mimc.Sum())

	mimc.Reset()

	// Amount,K, is in the merkle tree at index
	mp := merkle.MerkleProof{
		RootHash: c.Root,
		Path:     c.Path[:],
	}
	mp.VerifyProof(api, &mimc, c.Index)
	// Change == Amount - Withdrawal - Fee, and C, A, W, F are all non-negative
	// We express it by:
	// 		W <= A
	//		F <= A - W
	//		C = A - W - F
	api.AssertIsLessOrEqual(c.Withdrawal, c.Amount)
	api.AssertIsLessOrEqual(c.Fee, api.Sub(c.Amount, c.Withdrawal))
	api.AssertIsEqual(c.Change, api.Sub(c.Amount, c.Withdrawal, c.Fee))

	return nil
}
