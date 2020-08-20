package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

func MakeRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/wallets", allWallets).Methods(http.MethodGet)
	router.HandleFunc("/wallets/{walletID}", walletByID).Methods(http.MethodGet)
	router.HandleFunc("/transfer", allTransfers).Methods(http.MethodGet)
	router.HandleFunc("/transfer", transferFunds).Methods(http.MethodPost)

	return router
}
