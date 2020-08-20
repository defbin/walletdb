package web

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/defbin/walletdb/lib"
)

type walletResponse struct {
	ID       string `json:"id"`
	Balance  string `json:"balance"`
	Currency string `json:"currency"`
}

func walletToResponse(w *lib.Wallet) *walletResponse {
	return &walletResponse{
		ID:       w.ID.String(),
		Balance:  w.Balance.String(),
		Currency: w.Currency.String(),
	}
}

func allWallets(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("walletdb:db").(*sql.DB)

	wallets, err := lib.FindAllWallets(r.Context(), db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wr := make([]*walletResponse, len(wallets))
	for i := range wallets {
		wr[i] = walletToResponse(wallets[i])
	}

	j, err := json.Marshal(map[string][]*walletResponse{"data": wr})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		log.Printf("allWallets handler: %v\n", err.Error())
	}
}

func walletByID(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("walletdb:db").(*sql.DB)

	walletID, err := lib.ParseWalletID(mux.Vars(r)["walletID"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wlt, err := lib.FindWalletByID(r.Context(), db, walletID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if wlt == nil {
		http.NotFound(w, r)
		return
	}

	wr := walletToResponse(wlt)
	j, err := json.Marshal(wr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(j)
	if err != nil {
		log.Printf("walletByID handler: %v\n", err.Error())
	}
}
