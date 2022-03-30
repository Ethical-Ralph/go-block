package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"text/template"

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

func (ws *WalletServer) Transaction(w http.ResponseWriter, req *http.Request) {
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
		privateKey := utils.PublicKeyFromString(*t.SenderPrivateKey)
		value, err := strconv.ParseFloat(*t.Amount, 32)

		if err != nil {
			log.Println(err)
			io.WriteString(w, string(utils.JsonStatus("failed")))
			return
		}

		value32 := float32(value)
		fmt.Println(privateKey)
		fmt.Println(publicKey)
		fmt.Printf("%.1f\n", value32)

		fmt.Println(*t.RecipientBlockchainAddress)
		fmt.Println(*t.SenderBlockchainAddress)
		fmt.Println(*t.SenderPrivateKey)
		fmt.Println(*t.Amount)

		w.Header().Add("Content-Type", "application/json")

		io.WriteString(w, string(utils.JsonStatus("success")))

	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Invalid request method:", req.Method)
	}
}

func (ws *WalletServer) Run() {
	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/wallet", ws.Wallet)
	http.HandleFunc("/transaction", ws.Transaction)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(ws.Port())), nil))
}
