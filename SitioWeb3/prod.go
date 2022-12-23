package main

import (
	"fmt"
	"log"

	"github.com/mithorium/secure-fasthttp"
	"github.com/valyala/fasthttp"
)

func requestHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hello, world!\n")
}

func main() {
	secureMiddleware := secure.New(secure.Options{
		SSLRedirect: true,
		SSLHost:     "localhost:443", // This is optional in production. The default behavior is to just redirect the request to the HTTPS protocol. Example: http://github.com/some_page would be redirected to https://github.com/some_page.
	})

	secureHandler := secureMiddleware.Handler(requestHandler)

	// HTTP
	go func() {
		log.Fatal(fasthttp.ListenAndServe(":80", secureHandler))
	}()

	// HTTPS
	// To generate a development cert and key, run the following from your *nix terminal:
	// go run $GOROOT/src/pkg/crypto/tls/generate_cert.go --host="localhost"
	log.Fatal(fasthttp.ListenAndServeTLS(":443", "/etc/letsencrypt/live/www.redigo.cl/fullchain.pem", "/etc/letsencrypt/live/www.redigo.cl/privkey.pem", secureHandler))
}
