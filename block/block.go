package block

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ethical-Ralph/go-block/transaction"
)

type Block struct {
	Timestamp    int64
	Nonce        int
	PreviousHash string
	Transactions []*transaction.Transaction
	Hash         string
}

func NewBlock(nonce int, previousHash string, transactions []*transaction.Transaction) *Block {
	b := new(Block)
	b.Nonce = nonce
	b.PreviousHash = previousHash
	b.Timestamp = time.Now().UnixNano()
	b.Transactions = transactions
	b.Hash = b.CalculateHash()
	return b
}

func (b *Block) Print() {
	fmt.Printf("timestamp           %d\n", b.Timestamp)
	fmt.Printf("nonce               %d\n", b.Nonce)
	fmt.Printf("previousHash        %x\n", b.PreviousHash)
	fmt.Printf("hash                %x\n", b.Hash)
	for _, t := range b.Transactions {
		t.Print()
	}
}

func (b *Block) CalculateHash() string {
	jsonString, _ := b.MarshalJSON()
	return fmt.Sprintf("%x", sha256.Sum256([]byte(jsonString)))
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64                      `json:"timestamp"`
		Nonce        int                        `json:"nonce"`
		PreviousHash string                     `json:"previousHash"`
		Transactions []*transaction.Transaction `json:"transactions"`
	}{
		Timestamp:    b.Timestamp,
		Nonce:        b.Nonce,
		PreviousHash: fmt.Sprintf("%x", b.PreviousHash),
		Transactions: b.Transactions,
	})
}
