package site

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"bitbucket.org/pav5000/socketbot/config"
	"github.com/pkg/errors"
	"golang.org/x/crypto/acme/autocert"
)

const (
	dataDir = "./data/autocert"

	htmlIndex = `<!DOCTYPE html>
	<html>
	<head>
		<title>Welcome</title>
	</head>
	<body>
	<h1>Index page</h1><br><br>
	There will be a map here.
	</body>
	</html>`
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, htmlIndex)
}

func makeServerFromMux(mux *http.ServeMux) *http.Server {
	// set timeouts so that a slow or malicious client doesn't
	// hold resources forever
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
}

func makeHTTPServer() *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/", handleIndex)
	return makeServerFromMux(mux)

}

func makeHTTPToHTTPSRedirectServer() *http.Server {
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		newURI := "https://" + config.Config.HttpDomain + r.URL.String()
		http.Redirect(w, r, newURI, http.StatusFound)
	}
	mux := &http.ServeMux{}
	mux.HandleFunc("/", handleRedirect)
	return makeServerFromMux(mux)
}

func Run() error {
	if config.Config.HttpDomain == "" {
		fmt.Println("skipping https server start because http_domain is empty inconfig.yml")
		return nil
	}

	var m *autocert.Manager

	hostPolicy := func(ctx context.Context, host string) error {
		if host == config.Config.HttpDomain {
			return nil
		}
		return fmt.Errorf("only %s host is allowed", config.Config.HttpDomain)
	}

	err := os.MkdirAll(dataDir, 700)
	if err != nil {
		return errors.Wrap(err, "creating autocert data dir")
	}
	m = &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy,
		Cache:      autocert.DirCache(dataDir),
	}

	httpsSrv := makeHTTPServer()
	httpsSrv.Addr = config.Config.HttpsListen
	httpsSrv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

	go func() {
		fmt.Printf("Starting HTTPS server on %s\n", httpsSrv.Addr)
		err := httpsSrv.ListenAndServeTLS("", "")
		if err != nil {
			log.Fatalf("httpsSrv.ListendAndServeTLS() failed with %s", err)
		}
	}()

	var httpSrv *http.Server

	httpSrv = makeHTTPToHTTPSRedirectServer()
	// allow autocert handle Let's Encrypt callbacks over http
	if m != nil {
		httpSrv.Handler = m.HTTPHandler(httpSrv.Handler)
	}

	httpSrv.Addr = config.Config.HttpListen
	fmt.Printf("Starting HTTP server on %s\n", httpSrv.Addr)
	err = httpSrv.ListenAndServe()
	if err != nil {
		return errors.Wrap(err, "httpSrv.ListenAndServe()")
	}

	return nil
}
