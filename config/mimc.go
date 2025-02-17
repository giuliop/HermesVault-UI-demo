package config

import (
	"hash"

	"github.com/consensys/gnark-crypto/ecc"
	bls12_381mimc "github.com/consensys/gnark-crypto/ecc/bls12-381/fr/mimc"
	bn254mimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

// NewMimcF returns a MiMC hash function for the given curve.
func NewMimcF(curve ecc.ID) func(...[]byte) []byte {
	return func(data ...[]byte) []byte {
		return mimc(curve, data...)
	}
}

// mimc returns the hash of any number of byte slices.
// The byte slices are concatenated before passing them to the hasher.
// The total length must be a multiple of the curve block size or the hasher
// will panic.
// Each blocksize subslice of the concatenated inputs must represent a field
// element, if any represents a number greater than the field modulus, the
// hasher will panic.
func mimc(curve ecc.ID, data ...[]byte) []byte {
	var m hash.Hash
	switch curve {
	case ecc.BN254:
		m = bn254mimc.NewMiMC()
	case ecc.BLS12_381:
		m = bls12_381mimc.NewMiMC()
	}
	input := make([]byte, 0, 32*len(data))
	for _, slice := range data {
		input = append(input, slice...)
	}
	_, err := m.Write(input)
	if err != nil {
		panic(err)
	}
	return m.Sum(nil)
}
