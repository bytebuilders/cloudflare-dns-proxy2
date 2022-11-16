package main

import (
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	_ "github.com/cloudflare/cloudflare-go"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
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
func main() {
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
