package database

import (
	"context"
	"database/sql"

	pg "github.com/lib/pq"
	"github.com/zeebo/errs"
)

var (
	ErrWalletScan           = errs.Class("wallet scan failed")
	ErrCreateWallet         = errs.Class("create wallet")
	ErrFindWalletByID       = errs.Class("find wallet")
	ErrFindAllWallets       = errs.Class("find all wallets")
	ErrFindManyWalletsByIDs = errs.Class("find many wallets")
	ErrAddFunds             = errs.Class("add funds")
	ErrRemoveFunds          = errs.Class("remove funds")
)

type (
	WalletID = ID
	Currency = string
)

type Wallet interface {
	ID() WalletID
	Balance() Decimal
	Currency() Currency
}

type walletImpl struct {
	id       WalletID
	balance  Decimal
	currency Currency
}

func (w *walletImpl) ID() WalletID {
	return w.id
}

func (w *walletImpl) Balance() Decimal {
	return w.balance
}

func (w *walletImpl) Currency() Currency {
	return w.currency
}

const (
	createWalletQuery = `
	insert into wallets (balance, currency) values ($1, $2)
	returning id, balance, currency`
	findAllWalletsQuery      = `select id, balance, currency from wallets`
	findWalletByIDQuery      = `select id, balance, currency from wallets where id = $1`
	findManyWalletsByIDQuery = `select id, balance, currency from wallets where id = any($1)`
	incByAmountToWalletQuery = `
	update wallets set balance = balance + $1 where id = $2
	returning id, balance, currency`
	decByAmountToWalletQuery = `
	update wallets set balance = balance - $1 where id = $2
	returning id, balance, currency`
)

func scanWallet(s Scanner) (Wallet, error) {
	var w walletImpl

	err := s.Scan(&w.id, &w.balance, &w.currency)
	if err != nil {
		if errs.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, ErrWalletScan.Wrap(err)
	}

	return &w, nil
}

func CreateWallet(ctx context.Context, q ContextRowQuerier, balance Decimal, currency Currency) (Wallet, error) {
	w, err := scanWallet(q.QueryRowContext(ctx, createWalletQuery, balance, currency))
	if err != nil {
		return nil, ErrCreateWallet.Wrap(err)
	}

	return w, nil
}

func FindAllWallets(ctx context.Context, q ContextQuerier) ([]Wallet, error) {
	rows, err := q.QueryContext(ctx, findAllWalletsQuery)
	if err != nil {
		return nil, ErrFindAllWallets.Wrap(err)
	}
	defer rows.Close()

	var wallets []Wallet
	for rows.Next() {
		w, err := scanWallet(rows)
		if err != nil {
			return nil, ErrFindAllWallets.Wrap(err)
		}

		wallets = append(wallets, w)
	}

	if err = rows.Err(); err != nil {
		return nil, ErrFindAllWallets.Wrap(err)
	}

	return wallets, nil
}

func FindManyWalletsByIDs(ctx context.Context, q ContextQuerier, ids []WalletID) ([]Wallet, error) {
	arr := pg.Array(ids)
	rows, err := q.QueryContext(ctx, findManyWalletsByIDQuery, arr)
	if err != nil {
		return nil, ErrFindManyWalletsByIDs.Wrap(err)
	}
	defer rows.Close()

	var wallets []Wallet
	for rows.Next() {
		w, err := scanWallet(rows)
		if err != nil {
			return nil, ErrFindManyWalletsByIDs.Wrap(err)
		}

		wallets = append(wallets, w)
	}

	if err = rows.Err(); err != nil {
		return nil, ErrFindManyWalletsByIDs.Wrap(err)
	}

	return wallets, nil
}

func FindWalletByID(ctx context.Context, q ContextRowQuerier, id ID) (Wallet, error) {
	w, err := scanWallet(q.QueryRowContext(ctx, findWalletByIDQuery, id))
	if err != nil {
		return nil, ErrFindWalletByID.Wrap(err)
	}

	return w, nil
}

func AddFunds(ctx context.Context, q ContextRowQuerier, walletId WalletID, amount Decimal) (Wallet, error) {
	w, err := scanWallet(q.QueryRowContext(ctx, incByAmountToWalletQuery, amount, walletId))
	if err != nil {
		return nil, ErrAddFunds.Wrap(err)
	}

	return w, nil
}

func RemoveFunds(ctx context.Context, q ContextRowQuerier, walletId WalletID, amount Decimal) (Wallet, error) {
	w, err := scanWallet(q.QueryRowContext(ctx, decByAmountToWalletQuery, amount, walletId))
	if err != nil {
		return nil, ErrRemoveFunds.Wrap(err)
	}

	return w, nil
}
