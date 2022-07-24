package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	//"runtime"
	"syscall"
	"time"

	"github.com/dgrr/http2"
	"github.com/valyala/fasthttp"
)

type Config struct {
	Tiempo time.Duration `json:"Tiempo"`
}
type MyHandler struct {
	Conf  Config `json:"Conf"`
	Count uint64 `json:"Count"`
}

func main() {

	pass := &MyHandler{
		Conf: Config{},
	}

	s := &fasthttp.Server{
		Handler: pass.requestHandler,
		Name:    "http2 test",
	}

	http2.ConfigureServer(s, http2.ServerConfig{Debug: true})

	/*
		var port string
		if runtime.GOOS == "windows" {
			port = ":81"
		} else {
			port = ":80"
		}
	*/

	con := context.Background()
	con, cancel := context.WithCancel(con)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		for {
			select {
			case s := <-signalChan:
				switch s {
				case syscall.SIGHUP:
					pass.Conf.init()
				case os.Interrupt:
					cancel()
					os.Exit(1)
				}
			case <-con.Done():
				log.Printf("Done.")
				os.Exit(1)
			}
		}
	}()
	go func() {
		err := s.ListenAndServeTLS(":81", "", "")
		if err != nil {
			log.Fatalln(err)
		}
	}()
	if err := run(con, pass, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}

func (h *MyHandler) requestHandler(ctx *fasthttp.RequestCtx) {
	fmt.Printf("%s\n", ctx.Request.Header.Header())
	if ctx.Request.Header.IsPost() {
		fmt.Fprintf(ctx, "%s\n", ctx.Request.Body())
	} else {
		ctx.SetContentType("text/html; charset=utf-8")
		fmt.Fprintf(ctx, "<html><head><link rel='shortcut icon' type='image/x-icon' href='https://www.usinox.cl/favicon.ico'></head><body style='background:#f00'></body></html>")
	}
}

// DAEMON //
func (h *MyHandler) StartDaemon() {
	h.Conf.Tiempo = 100 * time.Second
	fmt.Println("DAEMON")
}
func (c *Config) init() {
	var tick = flag.Duration("tick", 1*time.Second, "Ticking interval")
	c.Tiempo = *tick
}
func run(con context.Context, c *MyHandler, stdout io.Writer) error {
	c.Conf.init()
	log.SetOutput(os.Stdout)
	for {
		select {
		case <-con.Done():
			return nil
		case <-time.Tick(c.Conf.Tiempo):
			c.StartDaemon()
		}
	}
}
