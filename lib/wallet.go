package lib

import (
	"context"

	"github.com/zeebo/errs"

	"github.com/defbin/walletdb/database"
)

var (
	ErrNewWalletFromDB      = errs.Class("new wallet from db")
	ErrCreateWallet         = errs.Class("create wallet")
	ErrFindAllWallets       = errs.Class("find all wallets")
	ErrFindManyWalletsByIDs = errs.Class("find many wallets by ids")
	ErrFindWalletByID       = errs.Class("find wallet by id")
	ErrAddFunds             = errs.Class("add funds")
	ErrRemoveFunds          = errs.Class("remove funds")
	ErrWalletDoesNotExist   = errs.Class("wallet does not exist")
)

type WalletID database.WalletID

func (id WalletID) String() string {
	return id.ToDB().String()
}

func (id WalletID) ToDB() database.WalletID {
	return database.WalletID(id)
}

func ParseWalletID(s string) (WalletID, error) {
	v, err := database.ParseID(s)
	return WalletID(v), err
}

func WalletIDFromDB(id database.WalletID) WalletID {
	return WalletID(id)
}

type Wallet struct {
	ID       WalletID
	Balance  Decimal
	Currency Currency
}

func NewWalletFromDB(wallet database.Wallet) (*Wallet, error) {
	if wallet == nil {
		return nil, nil
	}

	d, err := NewDecimalFromDB(wallet.Balance())
	if err != nil {
		return nil, ErrNewWalletFromDB.Wrap(err)
	}

	c, err := NewCurrency(wallet.Currency())
	if err != nil {
		return nil, ErrNewWalletFromDB.Wrap(err)
	}

	w := Wallet{
		ID:       WalletIDFromDB(wallet.ID()),
		Balance:  d,
		Currency: c,
	}

	return &w, nil
}

func CreateWallet(ctx context.Context, q database.ContextRowQueryExecutor, balance Decimal, currency Currency) (*Wallet, error) {
	w, err := database.CreateWallet(ctx, q, balance.ToDB(), currency.ToDB())
	if err != nil {
		return nil, ErrCreateWallet.Wrap(err)
	}

	rv, err := NewWalletFromDB(w)
	if err != nil {
		return nil, ErrCreateWallet.Wrap(err)
	}

	return rv, nil
}

func FindAllWallets(ctx context.Context, q database.ContextQuerier) ([]*Wallet, error) {
	ws, err := database.FindAllWallets(ctx, q)
	if err != nil {
		return nil, ErrFindAllWallets.Wrap(err)
	}

	rv := make([]*Wallet, len(ws))
	for i, w := range ws {
		rv[i], err = NewWalletFromDB(w)
		if err != nil {
			return nil, ErrFindAllWallets.Wrap(err)
		}
	}

	return rv, nil
}

func FindManyWalletsByIDs(ctx context.Context, q database.ContextQuerier, ids []WalletID) (map[WalletID]*Wallet, error) {
	dbIDs := make([]database.WalletID, len(ids))
	for i := range ids {
		dbIDs[i] = ids[i].ToDB()
	}

	ws, err := database.FindManyWalletsByIDs(ctx, q, dbIDs)
	if err != nil {
		return nil, ErrFindManyWalletsByIDs.Wrap(err)
	}

	rv := make(map[WalletID]*Wallet)
	for i := range ws {
		w, err := NewWalletFromDB(ws[i])
		if err != nil {
			return nil, ErrFindManyWalletsByIDs.Wrap(err)
		}

		rv[w.ID] = w
	}

	return rv, nil
}

func FindWalletByID(ctx context.Context, q database.ContextRowQueryExecutor, id WalletID) (*Wallet, error) {
	w, err := database.FindWalletByID(ctx, q, id.ToDB())
	if err != nil {
		return nil, ErrFindWalletByID.Wrap(err)
	}

	rv, err := NewWalletFromDB(w)
	if err != nil {
		return nil, ErrFindWalletByID.Wrap(err)
	}

	return rv, nil
}

func addFunds(ctx context.Context, q database.ContextRowQueryExecutor, wallet *Wallet, amount Decimal) (*Wallet, error) {
	w, err := database.AddFunds(ctx, q, wallet.ID.ToDB(), amount.ToDB())
	if err != nil {
		return nil, ErrAddFunds.Wrap(err)
	}

	wallet, err = NewWalletFromDB(w)
	if err != nil {
		return nil, ErrAddFunds.Wrap(err)
	}

	return wallet, nil
}

func removeFunds(ctx context.Context, q database.ContextRowQueryExecutor, wallet *Wallet, amount Decimal) (*Wallet, error) {
	w, err := database.RemoveFunds(ctx, q, wallet.ID.ToDB(), amount.ToDB())
	if err != nil {
		return nil, ErrRemoveFunds.Wrap(err)
	}

	wallet, err = NewWalletFromDB(w)
	if err != nil {
		return nil, ErrRemoveFunds.Wrap(err)
	}

	return wallet, nil
}
