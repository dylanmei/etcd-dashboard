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
	Self     *url.URL
	embedded *httputil.ReverseProxy
}

func NewProxy(self string) (*Proxy, error) {
	selfURL, err := url.Parse(self)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse self URL:", self)
	}

	leaderURL, err := ensureLeaderURL(self)
	if err != nil {
		return nil, fmt.Errorf("Couldn't determine leader: %v", err)
	}

	fmt.Printf("Starting reverse proxy. Using EtcD leader: %s\n", leaderURL.Host)
	return &Proxy{
		embedded: newLeaderProxy(leaderURL),
		Self:     selfURL,
	}, nil
}

func ensureLeaderURL(self string) (*url.URL, error) {
	resp, err := http.Get(self + "/v2/leader")
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

	statusCode := p.tryServe(p.embedded, w, r)
	switch statusCode {
	case 403, 500:
		fmt.Printf("Encountered trouble with leader: STATUS %d\n", statusCode)
	default:
		fmt.Printf("Done proxying request: STATUS %d\n", statusCode)
		return
	}

	leaderURL, err := ensureLeaderURL("http://"+p.Self.Host)
	if err != nil {
		fmt.Printf("Couldn't determine leader: %v\n", err)
		return
	}

	fmt.Printf("Found new leader: %s\n", leaderURL.Host)
	newProxy := newLeaderProxy(leaderURL)
	retryCode := p.tryServe(newProxy, w, r)

	if retryCode == 403 {
		fmt.Printf("Couldn't redirect request to leader: %s\n", leaderURL.Host)
	} else {
		p.embedded = newProxy
	}

	fmt.Printf("Done proxying request: STATUS %d\n", retryCode)
}

func (p *Proxy) tryServe(proxy *httputil.ReverseProxy, w http.ResponseWriter, r *http.Request) int {
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, r)

	copyHeader(w.Header(), rec.Header())
	w.WriteHeader(rec.Code)
	_, err := io.Copy(w, rec.Body)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	return rec.Code
}

func newLeaderProxy(leaderURL *url.URL) *httputil.ReverseProxy {
	p := httputil.NewSingleHostReverseProxy(leaderURL)
	p.Director = func(req *http.Request) {
		targetQuery := leaderURL.RawQuery

		req.URL.Scheme = leaderURL.Scheme
		req.URL.Host = leaderURL.Host
		req.Host = leaderURL.Host
		req.URL.Path = singleJoiningSlash(leaderURL.Path, req.URL.Path)

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}

	return p
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
