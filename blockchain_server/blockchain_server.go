package blockchainserver

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/Ethical-Ralph/go-block/blockchain"
	"github.com/Ethical-Ralph/go-block/transaction"
	"github.com/Ethical-Ralph/go-block/utils"
	"github.com/Ethical-Ralph/go-block/wallet"
)

var cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{port}
}

func (bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}

func (bcs *BlockchainServer) GetBlockchain() *blockchain.Blockchain {
	bc, ok := cache["blockchain"]
	if !ok {
		minersWallet := wallet.NewWallet()
		bc = blockchain.NewBlockchain(minersWallet.BlockchainAddress(), bcs.Port())
		cache["blockchain"] = bc
		log.Printf(("private_key %v"), minersWallet.PrivateKey())
		log.Printf(("public_key %v"), minersWallet.PublicKey())
		log.Printf(("blockchain_address %v"), minersWallet.BlockchainAddress())
	}
	return bc
}

func (bcs *BlockchainServer) GetChain(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockchain()
		m, _ := bc.MarshalJSON()
		io.WriteString(w, string(m))
	default:
		log.Printf("Invalid request method: %s", req.Method)
	}
}

func (bcs *BlockchainServer) Transactions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")

		bc := bcs.GetBlockchain()

		transactions := bc.TransactionPool()

		m, _ := json.Marshal(struct {
			Length       int                        `json:"length"`
			Transactions []*transaction.Transaction `json:"transactions"`
		}{
			Length:       len(transactions),
			Transactions: transactions,
		})
		io.WriteString(w, string(m[:]))
		return

	case http.MethodPost:
		{
			decoder := json.NewDecoder(req.Body)

			var t blockchain.TransactionRequest

			err := decoder.Decode(&t)
			if err != nil {
				log.Printf("ERROR: %v", err)
				io.WriteString(w, string(utils.JsonStatus("fail")))
				return
			}

			if !t.Validate() {
				log.Println("ERROR: missing fields")
				io.WriteString(w, string(utils.JsonStatus("field validation failed")))
				return
			}

			publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
			signature := utils.SignatureFromString(*t.Signature)

			bc := bcs.GetBlockchain()

			isCreated := bc.CreateTransaction(*t.SenderBlockchainAddress, *t.RecipientBlockchainAddress,
				*t.Value, publicKey, signature)

			w.Header().Add("Content-Type", "application/json")
			var m []byte

			if !isCreated {
				w.WriteHeader(http.StatusBadGateway)
				m = utils.JsonStatus("Fail")
			} else {
				w.WriteHeader(http.StatusCreated)
				m = utils.JsonStatus("success")
			}

			io.WriteString(w, string(m))
		}

	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Mine(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		isMined := bc.Mining()

		var m []byte

		if !isMined {
			w.WriteHeader(http.StatusBadRequest)
			m = utils.JsonStatus("fail")
		} else {
			m = utils.JsonStatus("success")
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m))
		return

	default:
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, string(utils.JsonStatus("Invalid HTTP method")))
	}
}

func (bcs *BlockchainServer) StartMine(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		bc.StartMining()

		w.Header().Add("Content-Type", "application/json")
		m := utils.JsonStatus("success")
		io.WriteString(w, string(m))
		return

	default:
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, string(utils.JsonStatus("Invalid HTTP method")))
	}
}

func (bcs *BlockchainServer) Amount(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:

		blockchainAdress := req.URL.Query().Get("blockchain_address")
		amount := bcs.GetBlockchain().CalculateTotalAmount(blockchainAdress)

		ar := blockchain.AmountResponse{Amount: amount}

		m, _ := ar.MarshalJSON()

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m[:]))
		return

	default:
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, string(utils.JsonStatus("Invalid HTTP method")))
	}
}

func (bcs *BlockchainServer) Start() {
	http.HandleFunc("/", bcs.GetChain)
	http.HandleFunc("/transaction", bcs.Transactions)
	http.HandleFunc("/mine", bcs.Mine)
	http.HandleFunc("/mine/start", bcs.StartMine)
	http.HandleFunc("/amount", bcs.Amount)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(bcs.Port())), nil))
}
