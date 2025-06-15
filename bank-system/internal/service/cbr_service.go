package service

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

type CBRService interface {
	GetKeyRate() (float64, error)
}

type cbrService struct {
	soapURL string
}

func NewCBRService() CBRService {
	return &cbrService{
		soapURL: "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx",
	}
}

func (s *cbrService) GetKeyRate() (float64, error) {
	soapRequest := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <soap:Body>
    <KeyRateXML xmlns="http://web.cbr.ru/" />
  </soap:Body>
</soap:Envelope>`

	req, err := http.NewRequest("POST", s.soapURL, bytes.NewBufferString(soapRequest))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRateXML")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(body); err != nil {
		return 0, err
	}

	keyRateElements := doc.FindElements("//KeyRate")

	if len(keyRateElements) == 0 {
		return 0, fmt.Errorf("key rate not found in response")
	}

	latestElement := keyRateElements[len(keyRateElements)-1]
	rateStr := latestElement.SelectElement("Rate").Text()

	rateStr = strings.Replace(rateStr, ",", ".", 1)
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return 0, err
	}

	return rate, nil
}
