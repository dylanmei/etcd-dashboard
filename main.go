package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"

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
	r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, path.Join("/mod/dashboard", req.URL.Path)+"/", 302)
	}))

	http.Handle("/", r)
	log.Println("Starting EtcD Dashboard")
	log.Println("Using EtcD Backend:", config.EtcdAddr)

	err := http.ListenAndServe(":"+config.ListenPort, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
