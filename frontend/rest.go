package frontend

import (
	"errors"
	"fmt"
	"html/template"
	"loanBackedBitcoin/core"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const (
	maxSystemPort       int           = 1023
	defaultPort         int           = 8080
	defaultAddress      string        = "127.0.0.1"
	defaultTimeout      time.Duration = 70 * time.Millisecond
	defaultTemplatePath               = "./static/html/*.html"
)

var (
	PortCannotBeZero error = errors.New("The configured Port cannot be zero!")
	InvalidAddress   error = errors.New("Make sure the address is on the format: 123.222.123.212")
)

type RestFrontend struct {
	port    int
	address string
	timeout time.Duration
	*template.Template
	*http.Server
	*core.Person
}

type RestFrontendOption func(*RestFrontend) error

func WithPort(port int) RestFrontendOption {

	return func(opt *RestFrontend) error {
		if port == 0 || port < maxSystemPort {
			return PortCannotBeZero
		}

		opt.port = port
		return nil
	}
}

func WithAddress(address string) RestFrontendOption {
	return func(opt *RestFrontend) error {
		if len(address) <= 1 {
			return InvalidAddress
		}

		opt.address = address
		return nil
	}
}

func WithIdleTimeout(timeout time.Duration) RestFrontendOption {
	return func(opt *RestFrontend) error {
		opt.timeout = timeout

		return nil
	}
}

func assignDefaultIfNil(rf *RestFrontend) error {
	if rf.port == 0 {
		rf.port = defaultPort
	}

	if len(rf.address) == 0 {
		rf.address = defaultAddress
	}

	if rf.timeout == 0 {
		rf.timeout = defaultTimeout
	}

	return nil
}

func (rf RestFrontend) calculate(rw http.ResponseWriter, r *http.Request) {
	var originalMonthlyValue float64
	var tenYearCounter int

	err := r.ParseForm()
	if err != nil {
		log.Println(err.Error())
		return
	}

	retirementAge, err := strconv.Atoi(r.Form.Get("retirement_age"))
	if err != nil {
		log.Println(err.Error())
		return
	}

	btcBalance, err := strconv.ParseFloat(r.Form.Get("btc_balance"), 64)
	if err != nil {
		log.Println(err.Error())
		return
	}

	fiatRetirementIncome, err := strconv.ParseFloat(r.Form.Get("fiat_retirement_income"), 64)
	if err != nil {
		log.Println(err.Error())
		return
	}

	btcExchangeRate, err := strconv.ParseFloat(r.Form.Get("btc_exchange_rate"), 64)
	if err != nil {
		log.Println(err.Error())
		return
	}
	btcGrowthRate, err := strconv.ParseFloat(r.Form.Get("btc_growth_rate"), 64)
	if err != nil {
		log.Println(err.Error())
		return
	}

	inflationRate, err := strconv.ParseFloat(r.Form.Get("fiat_inflation_rate"), 64)
	if err != nil {
		log.Println(err.Error())
		return
	}

	loanInterestRate, err := strconv.ParseFloat(r.Form.Get("loan_interest_rate"), 64)
	if err != nil {
		log.Println(err.Error())
		return
	}

	ltv, err := strconv.ParseFloat(r.Form.Get("ltv"), 64)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//TODO: This needs to be moved to main
	p := core.NewPerson(btcBalance, btcExchangeRate, inflationRate, 0, retirementAge, 0)
	originalMonthlyValue = fiatRetirementIncome / 12
	for range 100 - retirementAge {
		tenYearCounter++

		err := p.GetLoan(fiatRetirementIncome, ltv, loanInterestRate)
		if err != nil {
			log.Println()
			log.Println(err.Error())
			result := fmt.Sprintf("Your balance is over at age %d you cannot retire with $%f/month", p.Age, originalMonthlyValue)
			err = rf.ExecuteTemplate(rw, "result", result)
			if err != nil {
				log.Println(err.Error())
				return
			}

			return
		}

		p.ExchangeRate *= btcGrowthRate
		p.Age++
		p.PayLoan()
		fiatRetirementIncome *= inflationRate

		if tenYearCounter == 10 {
			tenYearCounter = 0
			btcGrowthRate = btcGrowthRate - 0.01
			log.Printf("\n\n THE BTC GROWTH RATE IS NOW: %f", btcGrowthRate)
		}
	}
	result := fmt.Sprintf("Congratulations you can retire at age %d with value $%.2f per month", retirementAge, originalMonthlyValue)
	if err := rf.ExecuteTemplate(rw, "result", result); err != nil {
		log.Println(err.Error())
		return
	}

}

func (rf *RestFrontend) Index(rw http.ResponseWriter, r *http.Request) {
	err := rf.ExecuteTemplate(rw, "index", nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func NewRestFrontend(opts ...RestFrontendOption) (*RestFrontend, error) {
	var restOpt RestFrontend

	for _, opt := range opts {
		err := opt(&restOpt)
		if err != nil {
			return nil, err
		}
	}

	templ, err := template.ParseGlob(defaultTemplatePath)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	restOpt.Template = templ

	err = assignDefaultIfNil(&restOpt)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	mux := mux.NewRouter()

	mux.HandleFunc("/", restOpt.Index).Methods("GET")
	mux.HandleFunc("/calculate", restOpt.calculate).Methods("POST")

	server := http.Server{
		IdleTimeout: restOpt.timeout,
		Addr:        fmt.Sprintf("%s:%d", restOpt.address, restOpt.port),
		Handler:     mux,
	}

	restOpt.Server = &server

	return &restOpt, nil
}

func (rf *RestFrontend) Start() error {
	if err := rf.Server.ListenAndServe(); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
