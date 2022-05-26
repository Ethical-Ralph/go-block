package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Ethical-Ralph/go-block/block"
	"github.com/Ethical-Ralph/go-block/transaction"
	"github.com/Ethical-Ralph/go-block/utils"
)

const (
	MINNING_DIFFICULTY = 3
	MINNING_SENDER     = "THE BLOCKCHAIN"
	MINNING_REWARD     = 1
	MINING_TIMER_SEC   = 20
)

type Blockchain struct {
	transactionPool   []*transaction.Transaction
	chain             []*block.Block
	blockchainAddress string
	port              uint16
	mux               sync.Mutex
}

func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {
	b := &block.Block{}
	bc := new(Blockchain)
	bc.blockchainAddress = blockchainAddress
	bc.port = port
	bc.CreateBlock(0, b.CalculateHash())
	return bc
}

func (bc *Blockchain) TransactionPool() []*transaction.Transaction {
	return bc.transactionPool
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			Block []*block.Block `json:"chains"`
		}{
			Block: bc.chain,
		})
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *block.Block {
	b := block.NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*transaction.Transaction{}
	return b
}

func (bc *Blockchain) CreateTransaction(sender string, recipient string, value float32,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	isTransacted := bc.AddTransaction(sender, recipient, value, senderPublicKey, s)

	// TODO
	// Sync

	return isTransacted
}

func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32, senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	t := transaction.NewTransaction(sender, recipient, value)

	if sender == MINNING_SENDER {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {
		// if bc.CalculateTotalAmount(sender) < value {
		// 	fmt.Println(bc.CalculateTotalAmount((sender)))
		// 	log.Panicln("ERROR: Wallet balance low")

		// }
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	} else {
		log.Panicln("ERROR: Verify Transaction")
	}
	return false
}

func (bc *Blockchain) VerifyTransactionSignature(
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *transaction.Transaction,
) bool {
	m, _ := t.MarshalJSON()
	h := sha256.Sum256(m)
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}

func (bc *Blockchain) Print() {
	// for i, txPool := range bc.transactionPool {
	// 	fmt.Printf("%s Transaction Pool %d %s \n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
	// 	txPool.Print()
	// 	fmt.Printf("%s\n", strings.Repeat("*", 55))
	// }
	for i, block := range bc.chain {
		fmt.Printf("%s Chain %d %s \n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
		fmt.Printf("%s\n", strings.Repeat("*", 55))
	}
}

func (bc *Blockchain) CopyTransactionPool() []*transaction.Transaction {
	transactions := make([]*transaction.Transaction, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions, transaction.NewTransaction(t.SenderBlockchainAddress, t.RecipientBlockchainAddress, t.Value))
	}
	return transactions
}

func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transaction []*transaction.Transaction, difficulty int) bool {
	zeros := strings.Repeat("0", difficulty)
	guessBlock := block.NewBlock(nonce, previousHash, transaction)
	guessHashStr := fmt.Sprintf("%x", guessBlock.CalculateHash())
	return guessHashStr[:difficulty] == zeros
}

func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactionPool()
	previousHash := bc.LastBlock().CalculateHash()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MINNING_DIFFICULTY) {
		nonce += 1
	}
	return nonce
}

func (bc *Blockchain) Mining() bool {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	if len(bc.transactionPool) == 0 {
		return false
	}

	bc.AddTransaction(MINNING_SENDER, bc.blockchainAddress, MINNING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().CalculateHash()
	bc.CreateBlock(nonce, previousHash)
	log.Println("action=mining, status=success")
	return true
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
}

func (bc *Blockchain) CalculateTotalAmount(senderAddress string) float32 {
	var total float32 = 0
	for _, chain := range bc.chain {
		for _, transaction := range chain.Transactions {
			if senderAddress == transaction.RecipientBlockchainAddress {
				total += transaction.Value
			}
			if senderAddress == transaction.SenderBlockchainAddress {
				total -= transaction.Value
			}
		}
	}
	return total
}

func (bc *Blockchain) LastBlock() *block.Block {
	return bc.chain[len(bc.chain)-1]
}

type TransactionRequest struct {
	SenderBlockchainAddress    *string  `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string  `json:"recipient_blockchain_address"`
	SenderPublicKey            *string  `json:"sender_public_key"`
	Value                      *float32 `json:"value"`
	Signature                  *string  `json:"signature"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderBlockchainAddress == nil || tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil || tr.Value == nil || tr.Signature == nil {
		return false
	}

	return true
}

type AmountResponse struct {
	Amount float32 `json:"amount"`
}

func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount float32 `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}
