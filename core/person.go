package core

import (
	"errors"
	// "log"
)

type Bank struct {
	Balance   float64
	Colateral float64
}

type Person struct {
	*BitcoinBalance
	MyBank        *Bank
	InflationRate float64
	FiatBalance   float64
	Age           int
	Debt          float64
}

type BitcoinBalance struct {
	Balance      float64
	ExchangeRate float64
}

type PersonOpt func(*Person) error

func NewPerson(btcBalance float64, exchangeRate float64, inflationRate float64, fiatBalance float64, age int, debt float64) *Person {
	var b Bank
	bBalance := BitcoinBalance{
		Balance:      btcBalance,
		ExchangeRate: exchangeRate,
	}

	person := Person{
		BitcoinBalance: &bBalance,
		MyBank:         &b,
		InflationRate:  inflationRate,
		FiatBalance:    fiatBalance,
		Age:            age,
		Debt:           debt,
	}

	return &person
}

func (p *Person) GetLoan(loanValue float64, ltv float64, loanInterestRate float64) error {
	var err error

	bankLoan := ltv * (loanValue / p.ExchangeRate)
	if bankLoan > p.Balance {
		err = errors.New("You cannot get a Loan, your bitcoin balance is lower than the value you would like to withdrawn on the loan.")
		return err
	}

	p.MyBank.Balance = bankLoan

	p.Balance -= bankLoan

	p.FiatBalance = loanValue

	p.Debt = loanValue * loanInterestRate

	// log.Println()
	// log.Println("#### Getting the loan ####")
	// log.Println("Your BTC Balance is now: ", p.balance)
	// log.Println("Your FIAT balance is now: ", p.fiatBalance)
	// log.Println("BTC price is: ", p.exchangeRate)
	// log.Println("You are this age: ", p.age)

	return err
}

func (p *Person) PayLoan() {

	bankBtcPayment := p.Debt / p.ExchangeRate
	payback := p.MyBank.Balance - bankBtcPayment

	p.FiatBalance = 0

	p.Balance += payback

	// log.Println("#### 12 Months Later ####")
	// log.Println("Your BTC Balance is now: ", p.balance)
	// log.Println("Your FIAT balance is now: ", p.fiatBalance)
	// log.Println("BTC price is: ", p.exchangeRate)
	// log.Println("You are this age: ", p.age)
}
