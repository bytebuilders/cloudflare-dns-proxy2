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
	"strings"
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
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		req.Header.Set("Authorization", "Bearer "+api.APIToken)
		// Accept-Encoding: gzip
		req.Header.Set("Accept-Encoding", "gzip")

		req.Host = "api.cloudflare.com"
	}

	frontendProxy := httptest.NewServer(&httputil.ReverseProxy{
		Director:  director,
		Transport: auther{},
	})
	defer frontendProxy.Close()

	fmt.Println(frontendProxy.URL)
	select {}

	//resp, err := http.Get(frontendProxy.URL)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//b, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Printf("%s", b)
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
