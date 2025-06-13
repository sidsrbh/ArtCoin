package structs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Block struct {
	Index        int
	Timestamp    string
	Hash         string
	PrevHash     string
	Transactions []Transaction
}

type TransactionType int

const (
	CoinTransfer TransactionType = iota
	ArtUpload
	ArtTransfer
	ArtUpdate
)

type TransactionStatus int

const (
	Pending TransactionStatus = iota
	Completed
	Confirmed
)

type Transaction struct {
	TransactionId string
	Type          TransactionType
	ArtID         string
	From          string
	To            string
	Amount        float64
	Fee           float64
	Signature     string
	ArtOwnership  ArtOwnership
	Status        TransactionStatus
}

type Blockchain struct {
	Blocks []*Block
	Mutex  *sync.Mutex
}

type ArtOwnership struct {
	Id            string            `json:"id"`
	ArtOwner      string            `json:"artOwner"`
	Price         float64           `json:"price"`
	Description   string            `json:"description"`
	Format        string            `json:"format"`
	Art           string            `json:"art"`           // ID or URL of the art media
	RelatedImages []string          `json:"relatedImages"` // IDs or URLs of related images
	RelatedVideos []string          `json:"relatedVideos"` // IDs or URLs of related videos
	ArtName       string            `json:"artName"`
	ArtLikes      int               `json:"artLikes"`
	ForSale       bool              `json:"forSale"`
	Thumbnail     string            `json:"thumbnail"` // ID or URL of the thumbnail
	Status        TransactionStatus `json:"status"`
}



type ArtOwnershipSummary struct {
	Id        string
	Thumbnail []byte
	ArtLikes  int
	ForSale   bool
	Price     float64
	Status    TransactionStatus
}

type ResponseMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Validator struct {
	Address string
	Stake   float64
}

func (tx *Transaction) Serialize() string {
	amount := strconv.FormatFloat(tx.Amount, 'f', 9, 64)
	fee := strconv.FormatFloat(tx.Fee, 'f', 9, 64)
	fields := []string{
		tx.TransactionId,
		strconv.Itoa(int(tx.Type)),
		tx.ArtID,
		tx.From,
		tx.To,
		amount,
		fee,
		tx.ArtOwnership.Id,
		strconv.Itoa(int(tx.Status)),
	}
	return strings.Join(fields, "|")
}

func (bc *Blockchain) calculateHash(block *Block) string {
	record := string(block.Index) + block.Timestamp + block.PrevHash
	for _, tx := range block.Transactions {
		record += tx.Serialize()
	}
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func (bc Blockchain) AddBlock(transactions []Transaction, vals []Validator) *Block {
	newBlock := &Block{
		Index:        len(bc.Blocks) + 1,
		Timestamp:    time.Now().String(),
		Transactions: transactions, // Add this line
	}
	newBlock.Hash = bc.calculateHash(newBlock)
	if len(bc.Blocks) > 0 {
		newBlock.PrevHash = bc.Blocks[len(bc.Blocks)-1].Hash
	}
	bc.Mutex.Lock()
	bc.Blocks = append(bc.Blocks, newBlock)
	bc.Mutex.Unlock()

	return newBlock
}

func (ao *ArtOwnership) IsArtOwnershipEmpty() bool {
	if ao.ArtOwner != "" {
		return false
	}
	if ao.Price != 0 {
		return false
	}
	if ao.Description != "" {
		return false
	}
	if ao.Format != "" {
		return false
	}
	if len(ao.Art) != 0 {
		return false
	}
	if len(ao.RelatedImages) != 0 {
		return false
	}
	if len(ao.RelatedVideos) != 0 {
		return false
	}
	if ao.ArtName != "" {
		return false
	}
	if ao.ArtLikes != 0 {
		return false
	}
	if ao.ForSale {
		return false
	}
	if len(ao.Thumbnail) != 0 {
		return false
	}
	fmt.Println("Art is empty")
	// Add more fields as needed
	return true
}

func (s TransactionStatus) String() string {
	return [...]string{"Pending", "Completed", "Confirmed"}[s]
}
