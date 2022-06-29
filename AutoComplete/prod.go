package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	lediscfg "github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"
	"github.com/valyala/fasthttp"
)

type MyHandler struct {
	Db           *ledis.DB         `json:"Db"`
	Conf         Config            `json:"Conf"`
	AutoComplete map[string]string `json:"Productos"`
	Letters      []int32           `json:"Letters"`
}
type Config struct {
	Tiempo time.Duration `json:"Tiempo"`
}

func main() {

	cfg := lediscfg.NewConfigDefault()

	var port string
	if runtime.GOOS == "windows" {
		port = ":81"
		cfg.DataDir = "C:/Go/LedisDB/AutoComplete"
	} else {
		port = ":80"
		cfg.DataDir = "/var/Go/LedisDB/AutoComplete"
	}

	l, _ := ledis.Open(cfg)
	db, _ := l.Select(0)

	pass := &MyHandler{
		Db:           db,
		AutoComplete: make(map[string]string, 0),
		Letters:      make([]int32, 0, 10),
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

func (h *MyHandler) SaveDb() {

	var data string = "HOLA MUNDO"
	h.AutoComplete["世界"] = data

	h.Letters = []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	/*
		var i, j uint32 = 0, 0
		for i = 1; i <= 100000; i++ {
			//h.AutoComplete[string(i)] = data
		}
		var m2 uint32 = 200000
		var buf []byte = []byte(data)
		for j = 100001; j <= m2; j++ {
			h.Db.Set(Int32tobytes(j), buf)
		}
	*/
}

func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {

	if string(ctx.Method()) == "GET" {
		switch string(ctx.Path()) {
		case "/auto":

			var lenstr int = ParamInt(ctx.QueryArgs().Peek("l"))
			var b strings.Builder

			if lenstr == 0 {
				b.Write([]byte{91})
				s, err := h.AutoComplete1(string(ctx.QueryArgs().Peek("p")))
				if err {
					b.Write(s)
				}
				b.Write([]byte{93})
				fmt.Fprint(ctx, b.String())
			} else {

			}

		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
}

func (h *MyHandler) AutoComplete1(str string) ([]byte, bool) {

	var Auto string
	var foundAuto bool
	if Auto, foundAuto = h.AutoComplete[str]; foundAuto {
		fmt.Println(Auto)
	} else {
		var val int = 0
		for i, v := range str {
			inArray(h.Letters, v)
		}
		val, _ := h.Db.Get(key)
		if len(val) > 0 {
			fmt.Println(val)
		}
	}
	return "", true

}

func Int32tobytes(i uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return Reverse(b)
}
func Reverse(numbers []uint8) []uint8 {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}
func ParamInt(data []byte) int {
	var x int
	for _, c := range data {
		x = x*10 + int(c-'0')
	}
	return x
}
func IntPow(n, m int) int {
	if m == 0 {
		return 1
	}
	result := n
	for i := 2; i <= m; i++ {
		result *= n
	}
	return result
}
func inArray(list []int32, v int32) int {
	for i, x := range list {
		if x == v {
			return i
		}
	}
	return -1
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

// DAEMON //

/*



rot13 := func(r rune) rune {
	switch {
	case r >= 'A' && r <= 'Z':
		return 'A' + (r-'A'+13)%26
	case r >= 'a' && r <= 'z':
		return 'a' + (r-'a'+13)%26
	}
	return r
}
fmt.Println(strings.Map(rot13, "'Twas brillig and the slithy gopher..."))



*/
