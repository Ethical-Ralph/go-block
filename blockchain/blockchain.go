package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/Ethical-Ralph/go-block/block"
	"github.com/Ethical-Ralph/go-block/transaction"
	"github.com/Ethical-Ralph/go-block/utils"
)

const (
	MINNING_DIFFICULTY = 1
	MINNING_SENDER     = "THE BLOCKCHAIN"
	MINNING_REWARD     = 2.1
)

type Blockchain struct {
	transactionPool   []*transaction.Transaction
	chain             []*block.Block
	blockchainAddress string
	port              uint16
}

func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {
	b := &block.Block{}
	bc := new(Blockchain)
	bc.blockchainAddress = blockchainAddress
	bc.port = port
	bc.CreateBlock(0, b.CalculateHash())
	return bc
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			Block []*block.Block `json:"chains"`
		}{
			Block: bc.chain,
		})
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash string) *block.Block {
	b := block.NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*transaction.Transaction{}
	return b
}

func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32, senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	t := transaction.NewTransaction(sender, recipient, value)

	if sender == MINNING_SENDER {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {
		if bc.CalculateTotalAmount(sender) < value {
			fmt.Println(bc.CalculateTotalAmount((sender)))
			log.Panicln("ERROR: Wallet balance low")

		}
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

func (bc *Blockchain) ValidProof(nonce int, previousHash string, transaction []*transaction.Transaction, difficulty int) bool {
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
	bc.AddTransaction(MINNING_SENDER, bc.blockchainAddress, MINNING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().CalculateHash()
	bc.CreateBlock(nonce, previousHash)
	log.Println("action=mining, status=success")
	return true
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
