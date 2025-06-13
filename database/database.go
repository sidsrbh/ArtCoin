package database

import (
	"fmt"
	"indicartcoin/sqldatabase"
	"indicartcoin/state"
	"indicartcoin/structs"
	"math"
	"math/rand"
	"sort"
	"sync"
)

const MaxTransactionsPerBlock = 5

// Initialize blockchain
var Blockchain = structs.Blockchain{
	Mutex:  &sync.Mutex{},
	Blocks: []*structs.Block{},
}
var AppState = &state.State{}

var UserDatabase map[string][]string

// Initialize validators
var Validators []structs.Validator

var PendingTransactions []structs.Transaction

var Transactions []structs.Transaction

var ArtSummary map[string]structs.ArtOwnershipSummary

func AddTransaction(tx structs.Transaction, blockchain *structs.Blockchain, vals []structs.Validator) {
	PendingTransactions = append(PendingTransactions, tx)
	if tx.Type == structs.ArtUpload {
		sqldatabase.AddArtOwnership(tx.ArtOwnership)
	}
	if tx.Type == structs.ArtUpdate {
		err := sqldatabase.UpdateArtOwnership(tx.ArtOwnership.Id, tx.ArtOwnership)
		if err != nil {
			fmt.Println("ArtOwnership Updated Succcessfully")
		}
	}
	//update sql database
	sqldatabase.AddPendingTransaction(tx)

	if len(PendingTransactions) >= MaxTransactionsPerBlock {
		newBlock := blockchain.AddBlock(PendingTransactions, vals)
		//update sql database
		sqldatabase.AddBlock(newBlock)
		finalizeValidation(vals, PendingTransactions)
		finalizeTransaction(PendingTransactions, newBlock)
		PendingTransactions = sqldatabase.LoadPendingTransactions() // Load New Pending Transactions
	}
}

func finalizeValidation(vals []structs.Validator, transactions []structs.Transaction) {
	// Calculate the total fees from all transactions
	totalFees := 0.0
	for _, tx := range transactions {
		totalFees += tx.Fee
	}

	// Sort validators by stake and a pinch of randomness
	sort.Slice(vals, func(i, j int) bool {
		randomFactorI := 1 + (rand.Float64()-0.5)/10.0
		randomFactorJ := 1 + (rand.Float64()-0.5)/10.0
		return float64(vals[i].Stake)*randomFactorI > float64(vals[j].Stake)*randomFactorJ
	})

	// Exponential decay parameters
	decayConstant := 0.5 // You can adjust this value

	// Calculate the normalization constant
	normalizationConstant := 0.0
	for i := 0; i < len(vals); i++ {
		normalizationConstant += math.Exp(-decayConstant * float64(i))
	}

	// Distribute rewards
	for i, _ := range vals {
		// Calculate reward using normalized exponential decay formula
		portion := math.Exp(-decayConstant*float64(i)) / normalizationConstant
		reward := totalFees * portion

		// Update validator reward
		AppState.Balances[vals[i].Address] += float64(reward)

		//update sql Database
		sqldatabase.UpdateBalance(vals[i].Address, float64(reward))
		sqldatabase.DeleteValidator(vals[i].Address)
	}
	Validators = sqldatabase.LoadValidators()
}

func finalizeTransaction(transactions []structs.Transaction, block *structs.Block) {
	for _, tx := range transactions {
		ApplyTransaction(tx, block)
	}
}

func ApplyTransaction(tx structs.Transaction, block *structs.Block) {

	switch tx.Type {
	case structs.ArtTransfer:
		// Transfer ownership of the art
		artownership, err := sqldatabase.FetchArtOwnershipByArtID(tx.ArtID)
		if err != nil {
			fmt.Println("An error occured: ", err.Error())
		}
		artownership.ArtOwner = tx.To
		artownership.Status = structs.Confirmed
		//update sql database
		sqldatabase.UpdateArtOwnership(tx.ArtID, *artownership)

		AppState.Balances[tx.From] += float64(tx.Amount)
		//update sql database
		sqldatabase.UpdateBalance(tx.From, AppState.Balances[tx.From])

		AppState.Balances[tx.To] -= float64(tx.Amount)
		//update sql database
		sqldatabase.UpdateBalance(tx.To, AppState.Balances[tx.To])
	case structs.ArtUpload:
		//AppState.ArtOwnership = sqldatabase.LoadArtOwnership()
		artownership, err := sqldatabase.FetchArtOwnershipByArtID(tx.ArtID)
		if err != nil {
			fmt.Println("An error occured: ", err.Error())
		}
		artownership.Status = structs.Confirmed
		//AppState.ArtOwnership[tx.ArtID] = artownership
		sqldatabase.UpdateArtOwnership(tx.ArtID, *artownership)
	case structs.ArtUpdate:
		artownership, err := sqldatabase.FetchArtOwnershipByArtID(tx.ArtID)
		if err != nil {
			fmt.Println("An error occured: ", err.Error())
		}
		artownership.Status = structs.Confirmed
		sqldatabase.UpdateArtOwnership(tx.ArtID, *artownership)

	case structs.CoinTransfer:
		AppState.Balances[tx.To] += tx.Amount
		AppState.Balances[tx.From] -= tx.Fee
		if AppState.Balances[tx.To] != AppState.Balances[tx.From] {
			AppState.Balances[tx.From] -= tx.Amount
		}

		sqldatabase.UpdateBalance(tx.To, AppState.Balances[tx.To])
		sqldatabase.UpdateBalance(tx.From, AppState.Balances[tx.From])
	}
	tx.Status = structs.Completed
	sqldatabase.AddTransaction(tx, block.Index)
	sqldatabase.DeletePendingTransaction(tx.TransactionId)

	// Update balances, rewards, etc. (if applicable)
}

func CreateBalanceTableEntry(publicKey string) {

}
