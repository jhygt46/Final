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

	"html/template"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type Config struct {
	Tiempo time.Duration `json:"Tiempo"`
}
type MyHandler struct {
	Conf  Config `json:"Conf"`
	Count uint64 `json:"Count"`
}
type Data struct {
	Id     int    `json:"Id"`
	Nombre string `json:"Nombre"`
}

var (
	imgHandler fasthttp.RequestHandler
	cssHandler fasthttp.RequestHandler
	jsHandler  fasthttp.RequestHandler
	port       string
)

func main() {

	if runtime.GOOS == "windows" {
		imgHandler = fasthttp.FSHandler("C:/Go/Final/SitioWeb1/img", 1)
		cssHandler = fasthttp.FSHandler("C:/Go/Final/SitioWeb1/css", 1)
		jsHandler = fasthttp.FSHandler("C:/Go/Final/SitioWeb1/js", 1)
		port = ":81"
	} else {
		imgHandler = fasthttp.FSHandler("/var/Pelao/img", 1)
		cssHandler = fasthttp.FSHandler("/var/Pelao/css", 1)
		jsHandler = fasthttp.FSHandler("/var/Pelao/js", 1)
		port = ":80"
	}

	pass := &MyHandler{Conf: Config{}}
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
		r := router.New()
		r.GET("/", Index)
		r.GET("/css/{name}", Css)
		r.GET("/js/{name}", Js)
		r.GET("/img/{name}", Img)
		r.GET("/json/{name}/{cant}", Json)
		fasthttp.ListenAndServe(port, r.Handler)
	}()
	if err := run(con, pass, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}

func Js(ctx *fasthttp.RequestCtx) {
	jsHandler(ctx)
}
func Css(ctx *fasthttp.RequestCtx) {
	cssHandler(ctx)
}
func Img(ctx *fasthttp.RequestCtx) {
	imgHandler(ctx)
}

func Index(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/html; charset=utf-8")
	t, err := TemplatePage("html/index.html")
	ErrorCheck(err)
	err = t.Execute(ctx, nil)
	ErrorCheck(err)
}
func Json(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")

	name := ctx.UserValue("name")
	cant := ctx.UserValue("cant")

	fmt.Println(name, cant)

	m1 := "[{\"Id\":1,\"Nombre\":\"Abc\"},{\"Id\":2,\"Nombre\":\"Eds\"}]"
	fmt.Fprintf(ctx, m1)
}

func ErrorCheck(e error) {
	if e != nil {
		fmt.Println("ERROR:", e)
	}
}
func TemplatePage(v string) (*template.Template, error) {
	t, err := template.ParseFiles(v)
	if err != nil {
		log.Print(err)
		return t, err
	}
	return t, nil
}

func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {

	if string(ctx.Method()) == "GET" {
		switch string(ctx.Path()) {
		case "/":
			h.Count++
			fmt.Fprintf(ctx, "HOLA1")
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
