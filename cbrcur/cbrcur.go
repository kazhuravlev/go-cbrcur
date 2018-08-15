package cbrcur

import (
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"context"
	"errors"
	"golang.org/x/net/html/charset"
	"net/url"
)

var (
	ErrInvalidConfiguration = errors.New("invalid configuration")
)

type Client struct {
	httpClient *http.Client
}

type Option func(*Client) error

func WithHttpClient(client *http.Client) Option {
	return func(c *Client) error {
		if client == nil {
			return ErrInvalidConfiguration
		}

		c.httpClient = client
		return nil
	}
}

func New(options ...Option) (*Client, error) {
	c := Client{
		httpClient: http.DefaultClient,
	}

	for _, option := range options {
		if err := option(&c); err != nil {
			return nil, err
		}
	}

	return &c, nil
}

func (c *Client) GetCurrencies(ctx context.Context) ([]Currency, error) {
	u := "http://www.cbr.ru/scripts/XML_valFull.asp"
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var currenciesResp currenciesResp
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&currenciesResp); err != nil {
		return nil, err
	}

	currencies := currenciesResp.Currencies
	for _, currency := range currencies {
		currency.ParentCode = strings.Trim(currency.ParentCode, " ")
	}

	return currencies, nil
}

func (c *Client) GetRatesReport(ctx context.Context, date *time.Time) (*Report, error) {
	query := url.Values{}
	if date != nil {
		query.Set("date_req", date.Format("02/01/2006"))
	}
	u := "http://www.cbr.ru/scripts/XML_daily.asp?" + query.Encode()

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var report reportResp
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	if err := decoder.Decode(&report); err != nil {
		return nil, err
	}

	reportDate, err := time.Parse("02.01.2006", report.Date)
	if err != nil {
		return nil, err
	}

	return &Report{
		Rates: report.Rates,
		Date:  reportDate,
	}, nil
}
