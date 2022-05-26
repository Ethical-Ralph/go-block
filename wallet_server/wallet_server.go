package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/Ethical-Ralph/go-block/blockchain"
	"github.com/Ethical-Ralph/go-block/utils"
	"github.com/Ethical-Ralph/go-block/wallet"
)

type WalletServer struct {
	port    uint16
	gateway string
}

func NewWalletServer(port uint16, gateway string) *WalletServer {
	return &WalletServer{port, gateway}
}

func (ws *WalletServer) Port() uint16 {
	return ws.port
}

func (ws *WalletServer) Gateway() string {
	return ws.gateway
}

func (ws *WalletServer) Index(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		t, _ := template.ParseFiles("wallet_server/templates/index.html")
		t.Execute(w, "")
	default:
		log.Printf("Invalid request method: %s", req.Method)
	}
}

func (ws *WalletServer) Wallet(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		w.Header().Add("Content-Type", "application/json")
		myWallet := wallet.NewWallet()
		m, _ := myWallet.MarshalJSON()
		io.WriteString(w, string(m))

	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Invalid request method:", req.Method)
	}
}

func (ws *WalletServer) CreateTransaction(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		w.Header().Add("Content-Type", "application/json")
		decoder := json.NewDecoder(req.Body)
		var t wallet.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Println(err)
			io.WriteString(w, string(utils.JsonStatus("failed")))
			return
		}

		if !t.Validate() {
			io.WriteString(w, string(utils.JsonStatus("failed: missing fields")))
			return
		}

		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		privateKey := utils.PrivateKeyFromString(*t.SenderPrivateKey, publicKey)
		value, err := strconv.ParseFloat(*t.Amount, 32)

		if err != nil {
			log.Println(err)
			io.WriteString(w, string(utils.JsonStatus("failed")))
			return
		}

		value32 := float32(value)
		// fmt.Println(privateKey)
		// fmt.Println(publicKey)
		// fmt.Printf("%.1f\n", value32)

		// fmt.Println(*t.RecipientBlockchainAddress)
		// fmt.Println(*t.SenderBlockchainAddress)
		// fmt.Println(*t.SenderPrivateKey)
		// fmt.Println(*t.Amount)

		w.Header().Add("Content-Type", "application/json")

		transaction := wallet.NewTransaction(privateKey, publicKey, *t.SenderBlockchainAddress, *t.RecipientBlockchainAddress, value32)

		signature := transaction.GenerateSignature()

		signatureStr := signature.String()

		bt := &blockchain.TransactionRequest{
			t.SenderBlockchainAddress,
			t.RecipientBlockchainAddress,
			t.SenderPublicKey,
			&value32,
			&signatureStr,
		}

		m, _ := json.Marshal(bt)

		buf := bytes.NewBuffer(m)

		resp, err := http.Post(ws.Gateway()+"/transaction", "application/json", buf)

		if err != nil {
			log.Println(err)
			io.WriteString(w, string(utils.JsonStatus(err.Error())))
			return
		}

		if resp.StatusCode == 201 {
			io.WriteString(w, string(utils.JsonStatus("success")))
			return
		}

		io.WriteString(w, string(utils.JsonStatus("fail")))

	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Invalid request method:", req.Method)
	}
}

func (ws *WalletServer) WalletAmount(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		blockchainAddress := req.URL.Query().Get("blockchain_address")

		endpoint := fmt.Sprintf("%s/amount", ws.Gateway())

		client := &http.Client{}
		bcsReq, _ := http.NewRequest("GET", endpoint, nil)

		q := bcsReq.URL.Query()
		q.Add("blockchain_address", blockchainAddress)
		bcsReq.URL.RawQuery = q.Encode()

		bcsResp, err := client.Do(bcsReq)

		if err != nil {
			log.Printf("ERROR: %v", err)
			io.WriteString(w, string(utils.JsonStatus(err.Error())))
			return
		}

		if bcsResp.StatusCode == 200 {
			decoder := json.NewDecoder(bcsResp.Body)
			var bar blockchain.AmountResponse
			err := decoder.Decode(&bar)

			if err != nil {
				log.Printf("ERROR: %v", err)
				io.WriteString(w, string(utils.JsonStatus(err.Error())))
				return
			}

			m, _ := json.Marshal(struct {
				Message string  `json:"message"`
				Amount  float32 `json:"amount"`
			}{
				Message: "success",
				Amount:  bar.Amount,
			})

			io.WriteString(w, string(m[:]))
		} else {
			io.WriteString(w, string(utils.JsonStatus(err.Error())))

		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Invalid request method:", req.Method)
	}
}

func (ws *WalletServer) Run() {
	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/wallet", ws.Wallet)
	http.HandleFunc("/wallet/amount", ws.WalletAmount)
	http.HandleFunc("/transaction", ws.CreateTransaction)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(ws.Port())), nil))
}
