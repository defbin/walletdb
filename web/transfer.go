package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/defbin/walletdb/lib"
)

type transferBody struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
}

type transferResponse struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    string    `json:"amount"`
	FeeAmount string    `json:"fee_amount"`
	Time      time.Time `json:"time"`
}

func transferToResponse(t *lib.Transfer) *transferResponse {
	return &transferResponse{
		From:      t.From.String(),
		To:        t.To.String(),
		Amount:    t.Amount.String(),
		FeeAmount: t.FeeAmount.String(),
		Time:      t.CreatedAt,
	}
}

func getServiceFee() lib.Decimal {
	// todo: make configurable
	const serviceFee = 1.5
	return lib.NewDecimal(serviceFee)
}

func allTransfers(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("walletdb:db").(*sql.DB)

	transfers, err := lib.FindAllTransfers(r.Context(), db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	wr := make([]*transferResponse, len(transfers))
	for i := range transfers {
		wr[i] = transferToResponse(transfers[i])
	}

	j, err := json.Marshal(map[string][]*transferResponse{"data": wr})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	_, err = w.Write(j)
	if err != nil {
		log.Printf("allWallets handler: %v\n", err.Error())
	}
}

func transferFunds(w http.ResponseWriter, r *http.Request) {
	var body transferBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	db := r.Context().Value("walletdb:db").(*sql.DB)

	from, to, err := parseWalletIDs(body.From, body.To)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	amount, err := lib.NewDecimalFromString(body.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	transfer, err := doTransfer(r.Context(), db, from, to, amount)
	if err != nil {
		var status int
		if lib.ErrTransferFunds.Has(err) {
			status = http.StatusBadRequest
		} else {
			status = http.StatusInternalServerError
		}

		http.Error(w, err.Error(), status)

		return
	}

	j, err := json.Marshal(transferToResponse(transfer))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	_, err = w.Write(j)
	if err != nil {
		log.Printf("transferFunds handler: %v\n", err.Error())
	}
}

func doTransfer(ctx context.Context, db *sql.DB, from, to lib.WalletID, amount lib.Decimal) (*lib.Transfer, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	params := lib.TransferFundsParams{
		From:   from,
		To:     to,
		Amount: amount,
		Fee:    getServiceFee(),
	}

	res, err := lib.TransferFunds(ctx, db, &params)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("transferFunds handler: %v", err)
		}

		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("transferFunds handler: %v", err)

		return nil, err
	}

	return res.Transfer(), nil
}

func parseWalletIDs(fromS, toS string) (from, to lib.WalletID, err error) {
	from, err = lib.ParseWalletID(fromS)
	if err != nil {
		return
	}

	to, err = lib.ParseWalletID(toS)

	return
}
