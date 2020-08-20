package lib

import (
	"context"
	"time"

	"github.com/zeebo/errs"

	"github.com/defbin/walletdb/database"
)

var (
	ErrNewTransferFromDB               = errs.Class("make transfer from db")
	ErrTransferFunds                   = errs.Class("transfer funds")
	ErrUnsupportedCurrencyConversation = ErrTransferFunds.New("unsupported currency conversation")
	ErrInsufficientFunds               = ErrTransferFunds.New("insufficient funds")
	errCreateTransfer                  = errs.Class("create transfer")
)

type TransferID database.TransferID

func (id TransferID) String() string {
	return id.ToDB().String()
}

func (id TransferID) ToDB() database.TransferID {
	return database.TransferID(id)
}

func ParseTransferID(s string) (TransferID, error) {
	v, err := database.ParseID(s)
	return TransferID(v), err
}

func TransferIDFomDB(id database.TransferID) TransferID {
	return TransferID(id)
}

type Transfer struct {
	ID        TransferID
	From      WalletID
	To        WalletID
	Amount    Decimal
	FeeAmount Decimal
	CreatedAt time.Time
}

func NewTransferFromDB(transfer database.Transfer) (*Transfer, error) {
	if transfer == nil {
		return nil, nil
	}

	amount, err := NewDecimalFromDB(transfer.Amount())
	if err != nil {
		return nil, ErrNewTransferFromDB.Wrap(err)
	}

	feeAmount, err := NewDecimalFromDB(transfer.FeeAmount())
	if err != nil {
		return nil, ErrNewTransferFromDB.Wrap(err)
	}

	t := Transfer{
		ID:        TransferIDFomDB(transfer.ID()),
		From:      WalletIDFromDB(transfer.From()),
		To:        WalletIDFromDB(transfer.To()),
		Amount:    amount,
		FeeAmount: feeAmount,
		CreatedAt: transfer.CreatedAt(),
	}

	return &t, nil
}

type TransferFundsParams struct {
	From   WalletID
	To     WalletID
	Amount Decimal
	Fee    Decimal
}

type TransferFundsResult interface {
	From() *Wallet
	To() *Wallet
	Transfer() *Transfer
}

type transferFundsResultImpl struct {
	from     *Wallet
	to       *Wallet
	transfer *Transfer
}

func (t *transferFundsResultImpl) From() *Wallet {
	return t.from
}

func (t *transferFundsResultImpl) To() *Wallet {
	return t.to
}

func (t *transferFundsResultImpl) Transfer() *Transfer {
	return t.transfer
}

func FindAllTransfers(ctx context.Context, q database.ContextQuerier) ([]*Transfer, error) {
	ws, err := database.FindAllTransfers(ctx, q)
	if err != nil {
		return nil, err
	}

	rv := make([]*Transfer, len(ws))
	for i, w := range ws {
		rv[i], err = NewTransferFromDB(w)
		if err != nil {
			return nil, err
		}
	}

	return rv, nil
}

func TransferFunds(ctx context.Context, q database.ContextQueryExecutor, params *TransferFundsParams) (TransferFundsResult, error) {
	ids := []WalletID{params.From, params.To}
	ws, err := FindManyWalletsByIDs(ctx, q, ids)
	if err != nil {
		return nil, ErrTransferFunds.Wrap(err)
	}

	from := ws[params.From]
	if from == nil {
		return nil, ErrTransferFunds.Wrap(ErrWalletDoesNotExist.New("%v", params.From))
	}
	to := ws[params.To]
	if to == nil {
		return nil, ErrTransferFunds.Wrap(ErrWalletDoesNotExist.New("%v", params.To))
	}

	if err := verifyWalletsBeforeTransfer(from, to, params.Amount, params.Fee); err != nil {
		return nil, ErrTransferFunds.Wrap(err)
	}

	transfer, err := doTransfer(ctx, q, from, to, params.Amount, params.Fee)
	if err != nil {
		return nil, ErrTransferFunds.Wrap(err)
	}

	return transfer, nil
}

func verifyWalletsBeforeTransfer(from, to *Wallet, amount, fee Decimal) error {
	if amount.Equal(NewDecimal(0)) || amount.Less(NewDecimal(0)) {
		return ErrTransferFunds.New("cannot transfer: %v", amount)
	}
	if !from.Currency.Equals(to.Currency) {
		return ErrUnsupportedCurrencyConversation
	}
	if from.Balance.Equal(NewDecimal(0)) {
		return ErrInsufficientFunds
	}

	feeAmount := calcFeeAmount(amount, fee)
	totalAmount := amount.Add(feeAmount)
	if from.Balance.Less(totalAmount) {
		return ErrInsufficientFunds
	}

	return nil
}

func doTransfer(ctx context.Context, q database.ContextRowQueryExecutor, from, to *Wallet, amount, fee Decimal) (TransferFundsResult, error) {
	feeAmount := calcFeeAmount(amount, fee)
	totalAmount := amount.Add(feeAmount)

	from, err := removeFunds(ctx, q, from, totalAmount)
	if err != nil {
		return nil, ErrTransferFunds.Wrap(err)
	}

	to, err = addFunds(ctx, q, to, amount)
	if err != nil {
		return nil, ErrTransferFunds.Wrap(err)
	}

	transfer, err := createTransaction(ctx, q, from, to, amount, feeAmount)
	if err != nil {
		return nil, ErrTransferFunds.Wrap(err)
	}

	rv := transferFundsResultImpl{
		from:     from,
		to:       to,
		transfer: transfer,
	}

	return &rv, nil
}

func calcFeeAmount(amount, fee Decimal) Decimal {
	return amount.Div(NewDecimalFromFloat(100)).Mul(fee)
}

func createTransaction(ctx context.Context, q database.ContextRowQueryExecutor, from, to *Wallet, amount, feeAmount Decimal) (*Transfer, error) {
	c, err := database.CreateTransaction(
		ctx,
		q,
		from.ID.ToDB(),
		to.ID.ToDB(),
		amount.ToDB(),
		feeAmount.ToDB(),
	)
	if err != nil {
		return nil, errCreateTransfer.Wrap(err)
	}

	t, err := NewTransferFromDB(c)
	if err != nil {
		return nil, errCreateTransfer.Wrap(err)
	}

	return t, nil
}
