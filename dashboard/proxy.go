package dashboard

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Proxy struct {
	To       *url.URL
	Leader   *url.URL
	embedded *httputil.ReverseProxy
}

func NewProxy(to string) (*Proxy, error) {
	toURL, err := url.Parse(to)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse target URL:", to)
	}

	leaderURL, err := ensureLeaderURL(to)
	if err != nil {
		return nil, fmt.Errorf("Couldn't determine leader: %v", err)
	}

	fmt.Printf("Starting reverse proxy. Using EtcD leader: %s\n", leaderURL.Host)
	proxy := &Proxy{
		embedded: httputil.NewSingleHostReverseProxy(leaderURL),
		To:       toURL,
		Leader:   leaderURL,
	}

	proxy.embedded.Director = proxy.director
	return proxy, nil
}

func ensureLeaderURL(to string) (*url.URL, error) {
	resp, err := http.Get(to + "/v2/leader")
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	leaderUrl, err := url.Parse(string(body))
	if err != nil {
		return nil, err
	}

	leaderHost, _, err := net.SplitHostPort(leaderUrl.Host)
	if err != nil {
		return nil, err
	}

	return url.Parse("http://" + leaderHost + ":4001")
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Start proxying request: %s %s\n", r.Method, r.RequestURI)

	rec := httptest.NewRecorder()
	p.embedded.ServeHTTP(rec, r)

	copyHeader(w.Header(), rec.Header())
	w.WriteHeader(rec.Code)
	_, err := io.Copy(w, rec.Body)
	if err != nil {
		fmt.Errorf("Error: %v", err)
	}

	fmt.Printf("Done proxying request: status %d\n", rec.Code)
}

func (p *Proxy) director(req *http.Request) {
	target := p.Leader
	targetQuery := target.RawQuery

	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.Host = target.Host
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)

	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
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

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
