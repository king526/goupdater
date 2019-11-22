package sdk

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

var (
	Version = "Undefined"
	Branch  = "Undefined"
	Commit  = "Undefined"
	modTime = "Undefined"
)

func init() {
	fi, err := os.Stat(os.Args[0])
	if err == nil {
		modTime = fi.ModTime().Format("2006-01-02 15:04:05")
	}
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Printf("Version   : %s\nBranch    : %s\nCommit    : %s\nModTime   : %s\n",
			Version, Branch, Commit, modTime)
		return
	}
}

func H2c(g *grpc.Server, h http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			g.ServeHTTP(w, r)
		} else if h != nil {
			h.ServeHTTP(w, r)
		} else {
			http.DefaultServeMux.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

func ListenAndServe(network, addr string) error {
	lis, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	RegisterUpgradeService(s)
	return s.Serve(lis)
}
