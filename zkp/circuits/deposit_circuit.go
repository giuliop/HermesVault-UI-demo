// Package circuits defines the zk-circuits for the application
package circuits

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

type DepositCircuit struct {
	Amount     frontend.Variable `gnark:",public"`
	Commitment frontend.Variable `gnark:",public"`
	K          frontend.Variable
	R          frontend.Variable
}

func (c *DepositCircuit) Define(api frontend.API) error {
	mimc, _ := mimc.NewMiMC(api)

	// hash(hash(Amount, K, R)) == Commitment
	mimc.Write(c.Amount)
	mimc.Write(c.K)
	mimc.Write(c.R)
	h := mimc.Sum()

	mimc.Reset()

	mimc.Write(h)
	api.AssertIsEqual(c.Commitment, mimc.Sum())

	return nil
}
