package main

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	_ "github.com/cloudflare/cloudflare-go"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

type auther struct {
	// rt http.RoundTripper
}

func (a auther) RoundTrip(req *http.Request) (*http.Response, error) {
	if data, err := httputil.DumpRequestOut(req, true); err == nil {
		fmt.Println("REQUEST: >>>>>>>>>>>>>>>>>>>>>>>")
		fmt.Println(string(data))
	}
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	if data, err := httputil.DumpResponse(resp, true); err == nil {
		fmt.Println("RESPONSE: >>>>>>>>>>>>>>>>>>>>>>>")
		fmt.Println(string(data))
	}
	return resp, nil
}

// https://pkg.go.dev/net/http/httputil#ReverseProxy
func main__() {
	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	target, _ := url.Parse(api.BaseURL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	d := proxy.Director
	proxy.Director = func(req *http.Request) {
		d(req)
		req.Host = ""
		req.Header.Set("Authorization", "Bearer "+api.APIToken)
	}
	frontendProxy := httptest.NewServer(proxy)
	defer frontendProxy.Close()

	fmt.Println(frontendProxy.URL)
	select {}
}

func main() {
	c, err := cloudflare.NewWithAPIToken(
		`e79163556a1557aaaa53fe501acca04d95d99287`,
		cloudflare.BaseURL("http://localhost:8000"))
	if err != nil {
		panic(err)
	}

	zones, err := c.ListZones(context.TODO())
	if err != nil {
		panic(err)
	}
	for _, z := range zones {
		fmt.Println(z.Name)

		_, e2 := c.CreateDNSRecord(context.TODO(), z.ID, cloudflare.DNSRecord{
			CreatedOn:  time.Time{},
			ModifiedOn: time.Time{},
			Type:       "A",
			Name:       "test2.bytebuilders.xyz",
			Content:    "69.164.204.85",
			Meta:       nil,
			Data:       nil,
			ID:         "",
			ZoneID:     z.ID,
			ZoneName:   "",
			Priority:   nil,
			TTL:        0,
			Proxied:    nil,
			Proxiable:  false,
			Locked:     false,
		})
		if e2 != nil {
			panic(e2)
		}
	}

}
