package state

import (
	"errors"
	"indicartcoin/blockchain"
	"indicartcoin/structs"
)

type State struct {
	Balances     map[string]float64              // Account balances
	ArtOwnership map[string]structs.ArtOwnership // ArtID to Owner
}

func (s *State) IsValidTransaction(tx structs.Transaction) (bool, error) {
	isValid, err := blockchain.VerifySignature(tx.Serialize(), tx.Signature, tx.From)

	if !isValid || err != nil {
		return false, err
	}

	switch tx.Type {

	case structs.ArtUpload:
		if tx.ArtID != tx.ArtOwnership.Id {
			return false, errors.New("art Id not same in Transaction and Ownership")
		}
		if tx.To != tx.From {
			return false, errors.New("to and From Different in Art Upload")
		}
	case structs.CoinTransfer:
		if tx.ArtID != "" {
			return false, errors.New("art Id Entered in Coin Transfer Transaction: " + tx.ArtID)
		}
		if !tx.ArtOwnership.IsArtOwnershipEmpty() {
			return false, errors.New("art Ownership hold values in coin transfer")
		}
	case structs.ArtTransfer:
		// Check if the sender owns the art
		owner, ownsArt := s.ArtOwnership[tx.ArtID]
		if !ownsArt || owner.ArtOwner != tx.From {
			return false, errors.New("malicious Transaction")
		}
		// Additional checks can be added here (e.g., sufficient balance to pay transaction fees)
		balance, Exists := s.Balances[tx.From]
		if !Exists || float64(balance) < tx.Amount {
			return false, errors.New("balance not sufficient")
		}
	case structs.ArtUpdate:
		{
			if tx.ArtID != tx.ArtOwnership.Id {
				return false, errors.New("art Id not same in Transaction and Ownership")
			}
			if tx.To != tx.From {
				return false, errors.New("to and From Different in Art Upload")
			}
		}
	}
	return true, nil
}
