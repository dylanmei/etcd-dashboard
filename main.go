package main

import (
	"fmt"
	"log"
	"net/http"
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

	fmt.Println("Starting EtcD Dashboard. Using EtcD Backend:", config.EtcdAddr)

	r := mux.NewRouter()
	// Routes for our EtcD reverse-proxy
	proxy, err := dashboard.NewProxy("http://" + config.EtcdAddr)
	if err != nil {
		log.Fatal("Couldn't create proxy:", err)
	}

	r.PathPrefix("/v2").Handler(proxy)
	r.PathPrefix("/version").Handler(proxy)

	// Routes for our Dashboard UI
	r.PathPrefix("/mod").Handler(http.StripPrefix("/mod", dashboard.HTTPHandler()))
	r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, path.Join("/mod/dashboard", req.URL.Path)+"/", 302)
	}))

	http.Handle("/", r)
	err = http.ListenAndServe(":"+config.ListenPort, nil)
	if err != nil {
		log.Fatal("Couldn't start server:", err)
	}
}
