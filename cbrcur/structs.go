package cbrcur

import (
	"encoding/xml"
	"strconv"
	"strings"
	"time"
)

type Currency struct {
	ID          string `xml:"ID,attr"`
	Name        string `xml:"Name"`
	EngName     string `xml:"EngName"`
	Nominal     int    `xml:"Nominal"`
	ParentCode  string `xml:"ParentCode"`
	ISONumCode  int    `xml:"ISO_Num_Code"`
	ISOCharCode string `xml:"ISO_Char_Code"`
}

type currenciesResp struct {
	Currencies []Currency `xml:"Item"`
}

type Rate struct {
	ID       string      `xml:"ID,attr"`
	NumCode  int         `xml:"NumCode"`
	CharCode string      `xml:"CharCode"`
	Nominal  int         `xml:"Nominal"`
	Name     string      `xml:"Name"`
	Value    customFloat `xml:"Value"`
}

type customFloat float64

func (c *customFloat) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(strings.Replace(v, ",", ".", 1), 64)
	if err != nil {
		return err
	}

	*c = customFloat(f)
	return nil
}

type Report struct {
	Rates []Rate
	Date  time.Time
}

type reportResp struct {
	Rates []Rate `xml:"Valute"`
	Date  string `xml:"Date,attr"`
}
