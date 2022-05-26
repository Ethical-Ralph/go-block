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
	PreviousHash [32]byte
	Transactions []*transaction.Transaction
}

func NewBlock(nonce int, previousHash [32]byte, transactions []*transaction.Transaction) *Block {
	b := new(Block)
	b.Nonce = nonce
	b.PreviousHash = previousHash
	b.Timestamp = time.Now().UnixNano()
	b.Transactions = transactions
	return b
}

func (b *Block) Print() {
	fmt.Printf("timestamp           %d\n", b.Timestamp)
	fmt.Printf("nonce               %d\n", b.Nonce)
	fmt.Printf("previousHash        %x\n", b.PreviousHash)
	for _, t := range b.Transactions {
		t.Print()
	}
}

func (b *Block) CalculateHash() [32]byte {
	jsonString, _ := b.MarshalJSON()
	return sha256.Sum256([]byte(jsonString))
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
