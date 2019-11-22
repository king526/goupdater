package main

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/king526/goupdater/sdk"
	"google.golang.org/grpc"
)

func main() {
	sdk.RegisterCommand("ping", func(args []string) (s string, e error) {
		return strings.Join(args, ","), nil
	})
	s := grpc.NewServer()
	sdk.RegisterUpgradeService(s)
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		bs, _ := ioutil.ReadAll(r.Body)
		w.Write([]byte(r.URL.RawQuery))
		w.Write(bs)
	})
	http.ListenAndServe(":8081", sdk.H2c(s, http.DefaultServeMux))
}
