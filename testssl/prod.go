package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mithorium/secure-fasthttp"
	"github.com/valyala/fasthttp"

	lediscfg "github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"
)

type Config struct {
	Tiempo time.Duration `json:"Tiempo"`
}
type MyHandler struct {
	Dbs   []*ledis.DB `json:"Dbs"`
	Conf  Config      `json:"Conf"`
	Count uint64      `json:"Count"`
}

func main() {

	num := 5
	pass := &MyHandler{
		Dbs: make([]*ledis.DB, num),
	}

	for i := 0; i < num; i++ {
		cfg := lediscfg.NewConfigDefault()
		cfg.DataDir = fmt.Sprintf("/var/Go/LedisDB/Init-%v", i)
		l, _ := ledis.Open(cfg)
		db, _ := l.Select(0)
		pass.Dbs[i] = db
	}

	pass.SaveDb()

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

		secureMiddleware := secure.New(secure.Options{SSLRedirect: true})
		secureHandler := secureMiddleware.Handler(pass.HandleFastHTTP)
		fasthttp.ListenAndServeTLS(":444", "/etc/letsencrypt/live/www.redigo.cl/fullchain.pem", "/etc/letsencrypt/live/www.redigo.cl/privkey.pem", secureHandler)

	}()
	if err := run(con, pass, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}
func (h *MyHandler) SaveDb() {

	len := len(h.Dbs)
	key := make([]byte, 6)
	j := 0
	data := []byte{254, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 126}

	for i := 0; i < len; i++ {
		for k := 0; k < 100; k++ {
			key[j] = 7
			j += copy(key[j+1:], newIntToBytes(i)) + 1
			j += copy(key[j:], newIntToBytes(k))
			h.Dbs[i].Set(key[j:], data)
		}
	}

}
func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {

	if string(ctx.Method()) == "GET" {
		switch string(ctx.Path()) {
		case "/":

		case "/stats":
			fmt.Fprintf(ctx, "Count %v", h.Count)
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
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

func newIntToBytes(num int) []byte {

	b := make([]byte, 4)
	var r int = num % 16777216
	b[0] = uint8(num / 16777216)
	b[1] = uint8(r / 65536)
	r = r % 65536
	b[2], b[3] = uint8(r/256), uint8(r%256)

	if num < 256 {
		return b[3:]
	} else if num < 65536 {
		return b[2:]
	} else if num < 16777216 {
		return b[1:]
	} else {
		return b
	}
}
