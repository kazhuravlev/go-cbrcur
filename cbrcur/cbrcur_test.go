package cbrcur_test

import (
	"context"
	"fmt"
	"github.com/kazhuravlev/go-cbrcur/cbrcur"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
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
	client, _ := cbrcur.New()

	ctx1, cancel := context.WithCancel(context.Background())
	cancel()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{"With canceled context", args{ctx1}, context.Canceled},
		{"Normal request", args{ctx2}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	client, _ := cbrcur.New()

	ctx1, cancel := context.WithCancel(context.Background())
	cancel()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	date := time.Date(2015, 8, 22, 0, 0, 0, 0, time.UTC)

	type args struct {
		ctx  context.Context
		date *time.Time
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{"With canceled context", args{ctx1, &date}, context.Canceled},
		{"With nil date", args{ctx2, nil}, nil},
		{"With date", args{ctx2, &date}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	client, _ := cbrcur.New()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	currencies, err := client.GetCurrencies(ctx)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	fmt.Println(currencies[:3])
}

func ExampleClient_GetRatesReport() {
	client, _ := cbrcur.New()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	date := time.Date(2015, 8, 22, 0, 0, 0, 0, time.UTC)
	rates, err := client.GetRatesReport(ctx, &date)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	fmt.Println(rates.Rates[:1])
}
