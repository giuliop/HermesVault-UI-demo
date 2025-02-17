package avm

import (
	"context"
	"fmt"
	"log"

	"github.com/algorand/go-algorand-sdk/v2/client/kmd"
	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/algorand/go-algorand-sdk/v2/types"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
)

type kmdConfig struct {
	URL      string
	Token    string
	Wallet   string
	Password string
}

var devnetKmdConfig = kmdConfig{
	URL:      "http://localhost:4002",
	Token:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	Wallet:   "unencrypted-default-wallet",
	Password: "",
}

func init() {
}

func devnetAlgodClient() *algod.Client {
	algodClient, err := algod.MakeClient(
		"http://localhost:4001",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	)
	if err != nil {
		log.Fatalf("Failed to create algod client: %v", err)
	}
	return algodClient
}

// GetFundedAddress returns a random test address that is funded with 1000 algo
func GetFundedAccount() crypto.Account {
	// generate a new random account with the offical algorand go sdk
	account := crypto.GenerateAccount()
	address := account.Address.String()
	ensureFunded(address, 1000*1_000_000)
	return account
}

// EnsureFunded checks if the given address has at least min microalgos and if not,
// tops it up from the default account
func ensureFunded(address string, min uint64) error {
	algodClient := devnetAlgodClient()
	recipientAccount, err := algodClient.AccountInformation(address).Do(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get account information: %v", err)
	}
	if recipientAccount.Amount < uint64(min) {
		account := getDevNetDefaultAccount()
		sp, err := algodClient.SuggestedParams().Do(context.Background())
		if err != nil {
			log.Fatalf("failed to get suggested params: %v", err)
		}
		waitRounds := uint64(4)
		sp.LastRoundValid = sp.FirstRoundValid + types.Round(waitRounds)
		txn, err := transaction.MakePaymentTxn(account.Address.String(),
			address, min-recipientAccount.Amount, nil, types.ZeroAddress.String(), sp)
		if err != nil {
			log.Fatalf("failed to make payment txn: %v", err)
		}
		txid, stx, err := crypto.SignTransaction(account.PrivateKey, txn)
		if err != nil {
			return fmt.Errorf("failed to sign transaction: %v", err)
		}
		_, err = algodClient.SendRawTransaction(stx).Do(context.Background())
		if err != nil {
			return fmt.Errorf("failed to send transaction: %v", err)
		}
		_, err = transaction.WaitForConfirmation(algodClient, txid, waitRounds,
			context.Background())
		if err != nil {
			return fmt.Errorf("error waiting for confirmation:  %v", err)
		}
	}
	return nil
}

func getDevNetDefaultAccount() *crypto.Account {
	kmdConfig := devnetKmdConfig
	client, err := kmd.MakeClient(
		kmdConfig.URL,
		kmdConfig.Token,
	)
	if err != nil {
		log.Fatalf("Failed to create kmd client: %s", err)
	}

	resp, err := client.ListWallets()
	if err != nil {
		log.Fatalf("failed to list wallets: %+v", err)
	}

	var walletId string
	for _, wallet := range resp.Wallets {
		if wallet.Name == kmdConfig.Wallet {
			walletId = wallet.ID
		}
	}

	if walletId == "" {
		log.Fatalf("no wallet named %s", kmdConfig.Wallet)
	}

	whResp, err := client.InitWalletHandle(walletId, kmdConfig.Password)
	if err != nil {
		log.Fatalf("failed to init wallet handle: %+v", err)
	}

	addrResp, err := client.ListKeys(whResp.WalletHandleToken)
	if err != nil {
		log.Fatalf("failed to list keys: %+v", err)
	}

	var accts []crypto.Account
	for _, addr := range addrResp.Addresses {
		expResp, err := client.ExportKey(whResp.WalletHandleToken, kmdConfig.Password, addr)
		if err != nil {
			log.Fatalf("failed to export key: %+v", err)
		}

		acct, err := crypto.AccountFromPrivateKey(expResp.PrivateKey)
		if err != nil {
			log.Fatalf("failed to create account from private key: %+v", err)
		}

		accts = append(accts, acct)
	}

	return &accts[0]
}
