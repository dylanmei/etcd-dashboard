package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/dylanmei/etcd-dashboard/config"
	"github.com/dylanmei/etcd-dashboard/dashboard"
	"github.com/gorilla/mux"
)

func main() {
	var config = config.New()
	if err := config.LoadFlags(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.PathPrefix("/mod").Handler(http.StripPrefix("/mod", dashboard.HTTPHandler()))

	target, _ := url.Parse("http://" + config.EtcdAddr)
	p := httputil.NewSingleHostReverseProxy(target)
	r.PathPrefix("/v2").Handler(p)

	http.Handle("/", r)
	var addr string = config.Host + ":" + config.Port
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
