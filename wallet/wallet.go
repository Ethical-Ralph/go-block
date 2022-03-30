package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Ethical-Ralph/go-block/utils"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	blockchainAddress string
}

type Transaction struct {
	senderPrivateKey           *ecdsa.PrivateKey
	senderPublicKey            *ecdsa.PublicKey
	SenderBlockchainAddress    string
	RecipientBlockchainAddress string

	Value float32
}
type TransactionRequest struct {
	SenderPrivateKey           *string `json:"sender_private_key"`
	RecipientBlockchainAddress *string `json:"recipient_blockchain_address"`
	SenderBlockchainAddress    *string `json:"sender_blockchain_address"`
	Amount                     *string `json:"amount"`
	SenderPublicKey            *string `json:"sender_public_key"`
}

func NewWallet() *Wallet {
	// 1. creating ECDA private key (32 bytes) public key (64 bytes)
	w := new(Wallet)
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	w.privateKey = privateKey
	w.publicKey = &w.privateKey.PublicKey
	// 2. Perform SHA-256 hashing on the public key (32 bytes)
	h2 := sha256.New()
	h2.Write(w.privateKey.X.Bytes())
	h2.Write(w.privateKey.Y.Bytes())
	digest2 := h2.Sum(nil)
	// 3. perform RIPEMD-160 hashing on the result of SHA-256 (20 bytes)
	h3 := ripemd160.New()
	h3.Write(digest2)
	digest3 := h3.Sum(nil)
	// 4. Add version byte in front of RIPEMD-160 hash (0x00 for Main Network)
	vd4 := make([]byte, 21)
	vd4[0] = 0x00
	copy(vd4[1:], digest3[:])
	// 5. Perform SHA-256 hash on the extended RIPEMD-160 result
	h5 := sha256.New()
	h5.Write(vd4)
	digest5 := h5.Sum(nil)
	// 6. perform SHA-256 hash on the result of the previous SHA-256 hash
	h6 := sha256.New()
	h6.Write(digest5)
	digest6 := h5.Sum(nil)
	// 7. take the first 4 bytes of the second SHA-256 hash for checksum
	chSum := digest6[:4]
	// 8. add the 4 checksum bytes from 7 at the end of the extended RIPEND-160 hash from 4 (25 bytes)
	dc8 := make([]byte, 25)
	copy(dc8[:21], vd4[:])
	copy(dc8[21:], chSum[:])
	// 9. convert the result from byte string into base58
	address := base58.Encode(dc8)
	w.blockchainAddress = address
	return w
}

func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

func (w *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x", w.privateKey.D.Bytes())
}

func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return w.publicKey
}

func (w *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%064x%064x", w.publicKey.X.Bytes(), w.publicKey.Y.Bytes())
}

func (w *Wallet) BlockchainAddress() string {
	return w.blockchainAddress
}

func (w *Wallet) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PrivateKey        string `json:"privateKey"`
		PublicKey         string `json:"publicKey"`
		BlockchainAddress string `json:"blockchainAddress"`
	}{
		PrivateKey:        w.PrivateKeyStr(),
		PublicKey:         w.PublicKeyStr(),
		BlockchainAddress: w.BlockchainAddress(),
	})
}

func NewTransaction(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, sender string, recipient string, value float32) *Transaction {
	return &Transaction{privateKey, publicKey, sender, recipient, value}
}

func (t *Transaction) GenerateSignature() *utils.Signature {
	m, _ := t.MarshalJSON()
	h := sha256.Sum256(m)
	r, s, _ := ecdsa.Sign(rand.Reader, t.senderPrivateKey, h[:])
	return &utils.Signature{R: r, S: s}
}

func (t *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 40))
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SenderBlockchainAddress    string  `json:"senderBlockchainAddress"`
		RecipientBlockchainAddress string  `json:"recipientBlockchainAddress"`
		Value                      float32 `json:"value"`
	}{
		SenderBlockchainAddress:    t.SenderBlockchainAddress,
		RecipientBlockchainAddress: t.RecipientBlockchainAddress,
		Value:                      t.Value,
	})
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderPrivateKey == nil || tr.RecipientBlockchainAddress == nil || tr.SenderBlockchainAddress == nil || tr.Amount == nil {
		return false
	}
	return true
}
