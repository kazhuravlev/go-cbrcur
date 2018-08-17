package cbrcur_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kazhuravlev/go-cbrcur/cbrcur"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type roundTripper struct {
	status  int
	headers http.Header
	body    []byte
}

func (rt roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: rt.status,
		Body:       ioutil.NopCloser(bytes.NewBuffer(rt.body)),
		Header:     rt.headers,
	}, nil
}

func mockClientResponse(status int, headers http.Header, body []byte) *http.Client {
	rt := roundTripper{
		status:  status,
		headers: headers,
		body:    body,
	}

	return &http.Client{Transport: rt}
}

var (
	bodyCurrencies, _    = ioutil.ReadFile("./testdata/currencies.xml")
	mockedCurrencyClient = mockClientResponse(200, make(http.Header), bodyCurrencies)

	bodyRates, _      = ioutil.ReadFile("./testdata/rate_report.xml")
	mockedRatesClient = mockClientResponse(200, make(http.Header), bodyRates)
)

func TestNew(t *testing.T) {
	type args struct {
		options []cbrcur.Option
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{"Without options", args{}, nil},
		{"With default client", args{[]cbrcur.Option{
			cbrcur.WithHttpClient(http.DefaultClient),
		}}, nil},
		{"With custom client", args{[]cbrcur.Option{
			cbrcur.WithHttpClient(new(http.Client)),
		}}, nil},
		{"With nil client", args{[]cbrcur.Option{
			cbrcur.WithHttpClient(nil),
		}}, cbrcur.ErrInvalidConfiguration},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := cbrcur.New(tt.args.options...)
			if tt.err != nil {
				assert.NotNil(t, err)
				assert.Nil(t, client)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, client)
		})
	}
}

func TestClient_GetCurrencies(t *testing.T) {
	ctx1, cancel := context.WithCancel(context.Background())
	cancel()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel2()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name       string
		httpClient *http.Client
		args       args
		err        error
	}{
		{"With canceled context", http.DefaultClient, args{ctx1}, context.Canceled},
		{"Normal request", mockedCurrencyClient, args{ctx2}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := cbrcur.New(
				cbrcur.WithHttpClient(tt.httpClient),
			)

			currencies, err := client.GetCurrencies(tt.args.ctx)
			if tt.err != nil {
				assert.NotNil(t, err)
				assert.Nil(t, currencies)
				return
			}

			assert.Nil(t, err)
			assert.NotNil(t, currencies)
			assert.True(t, len(currencies) >= 1)

			var parsedISONumCode, parsedISOCharCode bool
			for _, currency := range currencies {
				assert.NotEmpty(t, currency.ID)
				assert.NotEmpty(t, currency.Name)
				assert.NotEmpty(t, currency.EngName)
				assert.NotEmpty(t, currency.Nominal)
				assert.NotEmpty(t, currency.ParentCode)

				if currency.ISONumCode != 0 {
					parsedISONumCode = true
				}
				if currency.ISOCharCode != "" {
					parsedISOCharCode = true
				}
			}
			assert.True(t, parsedISONumCode)
			assert.True(t, parsedISOCharCode)
		})
	}
}

func TestClient_GetRateReport(t *testing.T) {
	ctx1, cancel := context.WithCancel(context.Background())
	cancel()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel2()

	date := time.Date(2015, 8, 22, 0, 0, 0, 0, time.UTC)

	type args struct {
		ctx  context.Context
		date *time.Time
	}
	tests := []struct {
		name       string
		httpClient *http.Client
		args       args
		err        error
	}{
		{"With canceled context", http.DefaultClient, args{ctx1, &date}, context.Canceled},
		{"With nil date", mockedRatesClient, args{ctx2, nil}, nil},
		{"With date", mockedRatesClient, args{ctx2, &date}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := cbrcur.New(
				cbrcur.WithHttpClient(tt.httpClient),
			)
			report, err := client.GetRatesReport(tt.args.ctx, tt.args.date)
			if tt.err != nil {
				assert.NotNil(t, err)
				assert.Nil(t, report)
				return
			}

			assert.Nil(t, err)
			require.NotNil(t, report)

			assert.NotNil(t, report.Rates)
			for _, rate := range report.Rates {
				assert.NotEmpty(t, rate.ID)
				assert.NotEmpty(t, rate.NumCode)
				assert.NotEmpty(t, rate.CharCode)
				assert.NotEmpty(t, rate.Nominal)
				assert.NotEmpty(t, rate.Name)
				assert.NotEmpty(t, rate.Value)
			}
		})
	}
}

func ExampleNew() {
	_, _ = cbrcur.New(
		cbrcur.WithHttpClient(http.DefaultClient),
	)
}

func ExampleClient_GetCurrencies() {
	client, _ := cbrcur.New(cbrcur.WithHttpClient(mockedCurrencyClient))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	currencies, _ := client.GetCurrencies(ctx)
	fmt.Println(currencies[:1])
	// Output: [{R01010 Австралийский доллар Australian Dollar 1 R01010     36 AUD}]
}

func ExampleClient_GetRatesReport() {
	client, _ := cbrcur.New(cbrcur.WithHttpClient(mockedRatesClient))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	date := time.Date(2015, 8, 22, 0, 0, 0, 0, time.UTC)
	rates, _ := client.GetRatesReport(ctx, &date)
	fmt.Println(rates.Rates[:1])
	// Output: [{R01010 36 AUD 1 Австралийский доллар 49.9059}]
}
