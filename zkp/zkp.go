// package zkp includes functionalities to interact with the zero-knowledge proof system
package zkp

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/giuliop/algoplonk"
	"github.com/giuliop/algoplonk/utils"
)

// ZkArgs returns the zk proof and public inputs as abi encoded arguments from the given
// assignment and compiled circuit
func ZkArgs(assignment frontend.Circuit, cc *algoplonk.CompiledCircuit) ([][]byte, error) {
	verifiedProof, err := cc.Verify(assignment)
	if err != nil {
		return nil, fmt.Errorf("failed to verify proof: %v", err)
	}
	proof := algoplonk.MarshalProof(verifiedProof.Proof)
	publicInputs, err := algoplonk.MarshalPublicInputs(verifiedProof.Witness)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public inputs: %v", err)
	}
	zkArgs, err := utils.AbiEncodeProofAndPublicInputs(proof, publicInputs)
	if err != nil {
		return nil, fmt.Errorf("failed to abi encode proof and public inputs: %v", err)
	}
	return zkArgs, nil
}
