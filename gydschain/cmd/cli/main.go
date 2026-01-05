package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/gydschain/gydschain/internal/crypto"
	"github.com/gydschain/gydschain/internal/tx"
)

func main() {
	// Define commands
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "wallet":
		walletCmd()
	case "tx":
		txCmd()
	case "query":
		queryCmd()
	case "stake":
		stakeCmd()
	case "version":
		fmt.Println("GYDS Chain CLI v1.0.0")
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`
GYDS Chain CLI - Command Line Interface

Usage:
  gydscli <command> [arguments]

Commands:
  wallet    Wallet management (create, import, export, balance)
  tx        Transaction operations (send, status)
  query     Query blockchain data (block, tx, account)
  stake     Staking operations (delegate, undelegate, rewards)
  version   Show version information
  help      Show this help message

Examples:
  gydscli wallet create --name mywallet
  gydscli wallet balance --address gyds1...
  gydscli tx send --from mywallet --to gyds1... --amount 100 --asset GYDS
  gydscli query block --height 1000
  gydscli stake delegate --validator gyds1... --amount 1000
`)
}

func walletCmd() {
	walletFlags := flag.NewFlagSet("wallet", flag.ExitOnError)
	action := walletFlags.String("action", "", "Action: create, import, export, balance, list")
	name := walletFlags.String("name", "", "Wallet name")
	address := walletFlags.String("address", "", "Wallet address")
	mnemonic := walletFlags.String("mnemonic", "", "Mnemonic phrase for import")
	output := walletFlags.String("output", "", "Output file for export")
	
	if len(os.Args) < 3 {
		fmt.Println("Usage: gydscli wallet --action <action> [options]")
		return
	}
	
	walletFlags.Parse(os.Args[2:])

	switch *action {
	case "create":
		createWallet(*name)
	case "import":
		importWallet(*name, *mnemonic)
	case "export":
		exportWallet(*address, *output)
	case "balance":
		showBalance(*address)
	case "list":
		listWallets()
	default:
		fmt.Println("Unknown wallet action. Use: create, import, export, balance, list")
	}
}

func createWallet(name string) {
	if name == "" {
		name = "default"
	}

	wallet, err := crypto.NewWallet(name)
	if err != nil {
		fmt.Printf("Error creating wallet: %v\n", err)
		return
	}

	fmt.Println("✅ Wallet created successfully!")
	fmt.Printf("   Name: %s\n", name)
	fmt.Printf("   Address: %s\n", wallet.Address())
	fmt.Printf("   Public Key: %s\n", wallet.KeyPair.PublicKeyHex())
	fmt.Println("\n⚠️  Please backup your private key securely!")
	fmt.Printf("   Private Key: %s\n", wallet.KeyPair.PrivateKeyHex())
}

func importWallet(name, mnemonic string) {
	if mnemonic == "" {
		fmt.Println("Please provide a mnemonic with --mnemonic")
		return
	}

	wallet, err := crypto.NewWalletFromMnemonic(name, mnemonic, "")
	if err != nil {
		fmt.Printf("Error importing wallet: %v\n", err)
		return
	}

	fmt.Println("✅ Wallet imported successfully!")
	fmt.Printf("   Name: %s\n", name)
	fmt.Printf("   Address: %s\n", wallet.Address())
}

func exportWallet(address, output string) {
	fmt.Printf("Exporting wallet %s to %s\n", address, output)
	// Implementation would save wallet data to file
}

func showBalance(address string) {
	if address == "" {
		fmt.Println("Please provide an address with --address")
		return
	}

	// In production, this would query the RPC server
	fmt.Printf("Balance for %s:\n", crypto.ShortAddress(address))
	fmt.Println("   GYDS: 0.00000000")
	fmt.Println("   GYD:  0.00000000")
	fmt.Println("\nNote: Connect to a node to see actual balance")
}

func listWallets() {
	fmt.Println("Saved wallets:")
	fmt.Println("   (No wallets found - wallet storage not implemented)")
}

func txCmd() {
	txFlags := flag.NewFlagSet("tx", flag.ExitOnError)
	action := txFlags.String("action", "send", "Action: send, status")
	from := txFlags.String("from", "", "Sender address or wallet name")
	to := txFlags.String("to", "", "Recipient address")
	amount := txFlags.Uint64("amount", 0, "Amount to send")
	asset := txFlags.String("asset", "GYDS", "Asset: GYDS or GYD")
	hash := txFlags.String("hash", "", "Transaction hash for status")
	
	if len(os.Args) < 3 {
		fmt.Println("Usage: gydscli tx --action send --from <addr> --to <addr> --amount <n> --asset <GYDS|GYD>")
		return
	}
	
	txFlags.Parse(os.Args[2:])

	switch *action {
	case "send":
		sendTx(*from, *to, *amount, *asset)
	case "status":
		txStatus(*hash)
	default:
		fmt.Println("Unknown tx action. Use: send, status")
	}
}

func sendTx(from, to string, amount uint64, asset string) {
	if from == "" || to == "" || amount == 0 {
		fmt.Println("Please provide --from, --to, and --amount")
		return
	}

	transaction := tx.NewTransfer(from, to, amount, asset)
	transaction.SetFee(21000) // Default fee

	hash, _ := transaction.HashHex()

	data, _ := json.MarshalIndent(map[string]interface{}{
		"hash":   hash,
		"from":   from,
		"to":     to,
		"amount": amount,
		"asset":  asset,
		"fee":    transaction.Fee,
		"status": "pending",
	}, "", "  ")

	fmt.Println("📤 Transaction created:")
	fmt.Println(string(data))
	fmt.Println("\nNote: Transaction signing requires wallet private key")
}

func txStatus(hash string) {
	if hash == "" {
		fmt.Println("Please provide --hash")
		return
	}

	fmt.Printf("Transaction status for %s:\n", hash)
	fmt.Println("   Status: pending")
	fmt.Println("\nNote: Connect to a node to check actual status")
}

func queryCmd() {
	queryFlags := flag.NewFlagSet("query", flag.ExitOnError)
	queryType := queryFlags.String("type", "", "Query type: block, tx, account")
	height := queryFlags.Uint64("height", 0, "Block height")
	hash := queryFlags.String("hash", "", "Block or tx hash")
	address := queryFlags.String("address", "", "Account address")
	
	if len(os.Args) < 3 {
		fmt.Println("Usage: gydscli query --type <block|tx|account> [options]")
		return
	}
	
	queryFlags.Parse(os.Args[2:])

	switch *queryType {
	case "block":
		queryBlock(*height, *hash)
	case "tx":
		queryTx(*hash)
	case "account":
		queryAccount(*address)
	default:
		fmt.Println("Unknown query type. Use: block, tx, account")
	}
}

func queryBlock(height uint64, hash string) {
	fmt.Printf("Querying block (height: %d, hash: %s)\n", height, hash)
	fmt.Println("Note: Connect to a node to query blocks")
}

func queryTx(hash string) {
	fmt.Printf("Querying transaction: %s\n", hash)
	fmt.Println("Note: Connect to a node to query transactions")
}

func queryAccount(address string) {
	if address == "" {
		fmt.Println("Please provide --address")
		return
	}

	fmt.Printf("Account: %s\n", address)
	fmt.Println("Note: Connect to a node to query account")
}

func stakeCmd() {
	stakeFlags := flag.NewFlagSet("stake", flag.ExitOnError)
	action := stakeFlags.String("action", "", "Action: delegate, undelegate, rewards, validators")
	validator := stakeFlags.String("validator", "", "Validator address")
	amount := stakeFlags.Uint64("amount", 0, "Amount to stake")
	from := stakeFlags.String("from", "", "Delegator address")
	
	if len(os.Args) < 3 {
		fmt.Println("Usage: gydscli stake --action <delegate|undelegate|rewards|validators> [options]")
		return
	}
	
	stakeFlags.Parse(os.Args[2:])

	switch *action {
	case "delegate":
		delegate(*from, *validator, *amount)
	case "undelegate":
		undelegate(*from, *validator, *amount)
	case "rewards":
		showRewards(*from)
	case "validators":
		listValidators()
	default:
		fmt.Println("Unknown stake action. Use: delegate, undelegate, rewards, validators")
	}
}

func delegate(from, validator string, amount uint64) {
	fmt.Printf("Delegating %d GYDS from %s to validator %s\n", amount, from, validator)
	fmt.Println("Note: Connect to a node to perform delegation")
}

func undelegate(from, validator string, amount uint64) {
	fmt.Printf("Undelegating %d GYDS from validator %s\n", amount, validator)
	fmt.Println("Note: Unbonding period is 21 days")
}

func showRewards(address string) {
	fmt.Printf("Staking rewards for %s:\n", address)
	fmt.Println("   Pending rewards: 0 GYDS")
	fmt.Println("Note: Connect to a node to check rewards")
}

func listValidators() {
	fmt.Println("Active validators:")
	fmt.Println("   (No validators - connect to a node)")
}
