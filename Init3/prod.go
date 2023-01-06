package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/valyala/fasthttp"

	lediscfg "github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"
)

type Config struct {
	Tiempo time.Duration `json:"Tiempo"`
}
type MyHandler struct {
	Db     *ledis.DB   `json:"Db"`
	Dbs    []*ledis.DB `json:"Dbs"`
	Conf   Config      `json:"Conf"`
	Count  int         `json:"Count"`
	CantDb int         `json:"CantDb"`
}

func main() {

	var port string
	if runtime.GOOS == "windows" {
		port = ":81"
	} else {
		port = ":8080"
	}

	var pass *MyHandler
	tipo := 1
	if tipo == 0 {
		numdb := 5
		pass = &MyHandler{
			Count:  0,
			Dbs:    make([]*ledis.DB, numdb),
			CantDb: numdb,
		}
		pass.Db = LedisConfig(0)
		for i := 0; i < numdb; i++ {
			pass.Dbs[i] = LedisConfig(i + 1)
		}
	}
	if tipo == 1 {
		pass = &MyHandler{}
		pass.Db = LedisConfig(0)
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
		fasthttp.ListenAndServe(port, pass.HandleFastHTTP)
	}()
	if err := run(con, pass, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}

func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {

	if string(ctx.Method()) == "GET" {
		switch string(ctx.Path()) {
		case "/":

			ctx.Response.Header.Set("Content-Type", "application/json")

			p1 := ParamInt(ctx.QueryArgs().Peek("p1"))
			p2 := ParamInt(ctx.QueryArgs().Peek("p2"))
			pais := ParamUint8(ctx.QueryArgs().Peek("pais"))
			key := make([]byte, 8)
			var j int = 1
			key[0] = pais

			j += copy(key[j:], IntToBytesMin3(p1))
			j += copy(key[j:], IntToBytesMin3(p2))

			if h.CantDb > 0 {
				h.Count++
				db := h.Count % (h.CantDb + 1)
				if db == 0 {
					val, _ := h.Db.Get(key[:j])
					if len(val) > 0 {
						ctx.SetBody(val)
					}
				} else {
					val, _ := h.Dbs[db-1].Get(key[:j])
					if len(val) > 0 {
						ctx.SetBody(val)
					}
				}

			} else {
				val, _ := h.Db.Get(key[:j])
				if len(val) > 0 {
					ctx.SetBody(val)
				}
			}

		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
}
func (h *MyHandler) SaveDb() {

	len := len(h.Dbs)
	key := make([]byte, 8)
	var j int
	data := []byte{254, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 34, 234, 234, 123, 12, 32, 64, 126}

	var count1 int = 1000 / (len + 1)
	var count2 int = 100

	for i := 0; i < len; i++ {
		for k := 0; k < count1; k++ {
			for m := 0; m < count2; m++ {
				j = 0
				key[j] = 7
				j += copy(key[j+1:], IntToBytesMin3(k)) + 1
				j += copy(key[j:], IntToBytesMin3(m))
				h.Dbs[i].Set(key[:j], data)
			}
		}
	}

	for k := 0; k < count1; k++ {
		for m := 0; m < count2; m++ {
			j = 0
			key[j] = 7
			j += copy(key[j+1:], IntToBytesMin3(k)) + 1
			j += copy(key[j:], IntToBytesMin3(m))
			h.Db.Set(key[:j], data)
		}
	}

	fmt.Println("SAVE DB")
}

func LedisConfig(path int) *ledis.DB {
	cfg := lediscfg.NewConfigDefault()
	cfg.DataDir = fmt.Sprintf("/var/Go/LedisDB/Init-%v", path)
	l, _ := ledis.Open(cfg)
	db, _ := l.Select(0)
	return db
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

func IntToBytes(num int) []byte {

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
func IntToBytesMin3(num int) []byte {

	b := make([]byte, 4)
	var r int = num % 16777216
	b[0] = uint8(num / 16777216)
	b[1] = uint8(r / 65536)
	r = r % 65536
	b[2], b[3] = uint8(r/256), uint8(r%256)

	if num < 16777216 {
		return b[1:]
	} else {
		return b
	}
}
func ParamInt(data []byte) int {
	var x int
	for _, c := range data {
		x = x*10 + int(c-'0')
	}
	return x
}
func ParamUint8(data []byte) uint8 {
	var x uint8
	for _, c := range data {
		x = x*10 + uint8(c-'0')
	}
	return x
}
