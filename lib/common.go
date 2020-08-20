package lib

import (
	"math/big"

	"github.com/zeebo/errs"

	"github.com/defbin/walletdb/database"
)

var (
	ErrUnsupportedCurrency  = errs.Class("unsupported currency")
	ErrInvalidDecimalString = errs.Class("invalid decimal string")
)

const (
	BTC = "BTC"
	ETH = "ETH"
)

type Currency struct {
	code string
}

func NewCurrency(code string) (Currency, error) {
	if code != BTC && code != ETH {
		return Currency{}, ErrUnsupportedCurrency.New(code)
	}

	return Currency{code}, nil
}

func (c Currency) Equals(o Currency) bool {
	return c.code == o.code
}

func (c Currency) String() string {
	return c.code
}

func (c Currency) ToDB() database.Currency {
	return c.String()
}

type Decimal struct {
	v big.Float
}

func NewDecimal(x float64) Decimal {
	return Decimal{*big.NewFloat(x)}
}

func NewDecimalFromDB(value database.Decimal) (Decimal, error) {
	return NewDecimalFromString(string(value))
}

func NewDecimalFromFloat(v float64) Decimal {
	return Decimal{*big.NewFloat(v)}
}

func NewDecimalFromString(s string) (Decimal, error) {
	var d Decimal
	_, ok := d.v.SetString(s)
	if !ok {
		return d, ErrInvalidDecimalString.New(s)
	}

	return d, nil
}

func (d Decimal) Copy() Decimal {
	rv := Decimal{}
	rv.v.Copy(&d.v)
	return rv
}

func (d Decimal) ToDB() database.Decimal {
	return database.Decimal(d.String())
}

func (d Decimal) String() string {
	return d.v.String()
}

func (d Decimal) Add(m Decimal) Decimal {
	rv := d.Copy()
	rv.v.Add(&rv.v, &m.v)
	return rv
}

func (d Decimal) Mul(m Decimal) Decimal {
	rv := d.Copy()
	rv.v.Mul(&rv.v, &m.v)
	return rv
}

func (d Decimal) Div(m Decimal) Decimal {
	rv := d.Copy()
	rv.v.Quo(&rv.v, &m.v)
	return rv
}

//
// func (d Decimal) Mul(m Decimal) Decimal {
// 	d1, o1 := d.v, m.v
// 	return Decimal{*d1.Mul(&d1, &o1)}
// }
//
// func (d Decimal) Div(m Decimal) Decimal {
// 	d1, o1 := d.v, m.v
// 	return Decimal{*d1.Quo(&d1, &o1)}
// }

func (d Decimal) Equal(o Decimal) bool {
	return d.v.Cmp(&o.v) == 0
}

func (d Decimal) Less(o Decimal) bool {
	return d.v.Cmp(&o.v) == -1
}
