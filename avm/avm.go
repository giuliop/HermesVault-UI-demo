// package avm provides functionalities to interact with the smart contracts onchain
package avm

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"webapp/config"

	"github.com/algorand/go-algorand-sdk/v2/abi"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
)

type algodConfig struct {
	URL   string
	Token string
}

var client *algod.Client

func init() {
	algodConfig, err := readAlgodConfigFromDir(config.AlgodPath)
	if err != nil {
		log.Fatalf("failed to read algod config: %v", err)
	}
	client, err = algod.MakeClient(
		algodConfig.URL,
		algodConfig.Token,
	)
	if err != nil {
		log.Fatalf("Failed to create algod client: %v", err)
	}
}

func algodClient() *algod.Client {
	return client
}

func CompileTealFromFile(tealPath string) ([]byte, error) {
	algodClient := algodClient()

	teal, err := os.ReadFile(tealPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s from file: %v", tealPath, err)
	}

	result, err := algodClient.TealCompile(teal).Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to compile %s: %v", tealPath, err)
	}
	binary, err := base64.StdEncoding.DecodeString(result.Result)
	if err != nil {
		log.Fatalf("failed to decode approval program: %v", err)
	}

	return binary, nil
}

// abiEncode encodes arg into its abi []byte representation
func abiEncode(arg any, abiTypeName string) ([]byte, error) {
	abiType, err := abi.TypeOf(abiTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get abi type: %v", err)
	}
	abiArg, err := abiType.Encode(arg)
	if err != nil {
		return nil, fmt.Errorf("failed to encode noChange: %v", err)
	}
	return abiArg, nil
}

// readAlgodConfigFromDir reads the algod URL and token from the given directory
func readAlgodConfigFromDir(dir string) (*algodConfig, error) {
	urlPath := filepath.Join(dir, "algod.net")
	url, err := os.ReadFile(urlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read algod url: %v", err)
	}
	tokenPath := filepath.Join(dir, "algod.token")
	token, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read algod token: %v", err)
	}
	return &algodConfig{
		URL:   "http://" + strings.TrimSpace(string(url)),
		Token: strings.TrimSpace(string(token)),
	}, nil
}
