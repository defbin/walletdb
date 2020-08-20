package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeebo/errs"
)

var (
	ErrScanTransfer     = errs.Class("scan transfer")
	ErrCreateTransfer   = errs.Class("create transfer")
	ErrFindAllTransfers = errs.Class("find all transfers")
	ErrFindTransferByID = errs.Class("find transfer by id")
)

type TransferID = ID

type Transfer interface {
	ID() TransferID
	From() WalletID
	To() WalletID
	Amount() Decimal
	FeeAmount() Decimal
	CreatedAt() time.Time
}

type transactionImpl struct {
	id        TransferID
	from      WalletID
	to        WalletID
	amount    Decimal
	feeAmount Decimal
	createdAt time.Time
}

func (t *transactionImpl) ID() TransferID {
	return t.id
}

func (t *transactionImpl) From() WalletID {
	return t.from
}

func (t *transactionImpl) To() WalletID {
	return t.to
}

func (t *transactionImpl) Amount() Decimal {
	return t.amount
}

func (t *transactionImpl) FeeAmount() Decimal {
	return t.feeAmount
}

func (t *transactionImpl) CreatedAt() time.Time {
	return t.createdAt
}

const (
	createTransactionQuery = `
	insert into transactions (sender, receiver, amount, fee_amount) values ($1, $2, $3, $4)
	returning id, sender, receiver, amount, fee_amount, created_at
	`
	findAllTransfersQuery = `select id, sender, receiver, amount, fee_amount, created_at from transactions`
	findTransferByIDQuery = `select id, sender, receiver, amount, fee_amount, created_at from transactions where id = $1`
)

func scanTransfer(s Scanner) (Transfer, error) {
	var t transactionImpl

	err := s.Scan(&t.id, &t.from, &t.to, &t.amount, &t.feeAmount, &t.createdAt)
	if err != nil {
		if errs.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, ErrScanTransfer.Wrap(err)
	}

	return &t, nil
}

func CreateTransaction(ctx context.Context, q ContextRowQuerier, from, to WalletID, amount, feeAmount Decimal) (Transfer, error) {
	t, err := scanTransfer(q.QueryRowContext(ctx, createTransactionQuery, from, to, amount, feeAmount))
	if err != nil {
		return nil, ErrCreateTransfer.Wrap(err)
	}

	return t, nil
}

func FindAllTransfers(ctx context.Context, q ContextQuerier) ([]Transfer, error) {
	rows, err := q.QueryContext(ctx, findAllTransfersQuery)
	if err != nil {
		return nil, ErrFindAllTransfers.Wrap(err)
	}
	defer rows.Close()

	var transfers []Transfer
	for rows.Next() {
		w, err := scanTransfer(rows)
		if err != nil {
			return nil, ErrFindAllTransfers.Wrap(err)
		}

		transfers = append(transfers, w)
	}

	if err = rows.Err(); err != nil {
		return nil, ErrFindAllTransfers.Wrap(err)
	}

	return transfers, nil
}

func FindTransferByID(ctx context.Context, q ContextRowQuerier, id TransferID) (Transfer, error) {
	t, err := scanTransfer(q.QueryRowContext(ctx, findTransferByIDQuery, id))
	if err != nil {
		return nil, ErrFindTransferByID.Wrap(err)
	}

	return t, nil
}
