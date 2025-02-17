package avm

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"webapp/config"
	"webapp/models"

	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/giuliop/algoplonk/utils"
)

// appSetupDirPath is the path to the app setup directory
var appSetupDirPath = config.AppSetupDirPath

// the setup filenames
const (
	appFile                       = "App.json"
	appArc32File                  = "APP.arc32.json"
	tssTealFile                   = "TSS.tok"
	depositVerifierTealFile       = "DepositVerifier.tok"
	withdrawalVerifierTealFile    = "WithdrawalVerifier.tok"
	treeConfigFile                = "TreeConfig.json"
	compiledDepositCircuitFile    = "CompiledDepositCircuit.bin"
	compiledWithdrawalCircuitFile = "CompiledWithdrawalCircuit.bin"
)

// App is the global app instance
var App *models.App

type AppJson struct {
	Id            uint64 `json:"id"`
	CreationBlock uint64 `json:"creationBlock"`
}

func init() {
	App = setupApp()
}

// setupApp sets up the app instance from the app setup files
// It panics if the setup fails
func setupApp() *models.App {
	app := models.App{}
	appJson := AppJson{}

	decodeJSONFile(pathTo(appFile), &appJson)
	app.Id = appJson.Id
	decodeJSONFile(pathTo(appArc32File), &app.Schema)
	app.TSS = readlogicsig(pathTo(tssTealFile))
	app.DepositVerifier = readlogicsig(pathTo(depositVerifierTealFile))
	app.WithdrawalVerifier = readlogicsig(pathTo(withdrawalVerifierTealFile))
	app.TreeConfig = readTreeConfiguration(pathTo(treeConfigFile))

	var err error
	app.DepositCc, err = utils.DeserializeCompiledCircuit(pathTo(compiledDepositCircuitFile))
	if err != nil {
		log.Fatalf("Error deserializing compiled deposit circuit: %v", err)
	}
	app.WithdrawalCc, err = utils.DeserializeCompiledCircuit(pathTo(
		compiledWithdrawalCircuitFile))
	if err != nil {
		log.Fatalf("Error deserializing compiled withdrawal circuit: %v", err)
	}

	return &app
}

func readlogicsig(compiledPath string) *models.Lsig {
	bytecode, err := os.ReadFile(compiledPath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	lsigAccount, err := crypto.MakeLogicSigAccountEscrowChecked(bytecode, nil)
	if err != nil {
		log.Fatalf("Error creating  logic sig account: %v", err)
	}
	address, err := lsigAccount.Address()
	if err != nil {
		log.Fatalf("Error getting lsig address: %v", err)
	}
	return &models.Lsig{
		Account: lsigAccount,
		Address: address,
	}
}

// readTreeConfiguration reads the tree configuration from the given file
func readTreeConfiguration(treeConfigPath string) models.TreeConfig {
	treeConfig := models.TreeConfig{}
	decodeJSONFile(treeConfigPath, &treeConfig)
	treeConfig.HashFunc = config.Hash
	return treeConfig
}

// pathTo returns the path to the file in the artefacts directory
func pathTo(file string) string {
	return filepath.Join(appSetupDirPath, file)
}

// DecodeJSONFile decodes the JSON filepath into the given interface
func decodeJSONFile(filepath string, v interface{}) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Error opening file %s: %v", filepath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(v); err != nil {
		log.Fatalf("Error decoding file %s: %v", filepath, err)
	}
}
