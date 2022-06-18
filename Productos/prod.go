package main

import (
	"runtime"

	lediscfg "github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"
	"github.com/valyala/fasthttp"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type MyHandler struct {
	Db    *ledis.DB         `json:"Db"`
	Cuads map[uint32][]byte `json:"Cuads"`
}
type Params struct {
	C []uint32 `json:"C"` // CUADRANTES
}

func main() {

	cfg := lediscfg.NewConfigDefault()

	var port string
	if runtime.GOOS == "windows" {
		port = ":81"
		cfg.DataDir = "C:/Go/LedisDB"
	} else {
		port = ":80"
		cfg.DataDir = "/var/Go/LedisDB"
	}

	l, _ := ledis.Open(cfg)
	db, _ := l.Select(0)

	h := &MyHandler{
		Db: db,
	}

	fasthttp.ListenAndServe(port, h.HandleFastHTTP)

}

func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {

	if string(ctx.Method()) == "GET" {
		switch string(ctx.Path()) {
		case "/ProdDesc":
			val, _ := h.Db.Get(ctx.QueryArgs().Peek("c"))
			if len(val) > 0 {
				var CantProd, w int
				CantProd, w = DecodeSpecialBytes(val[0:2], 200)
			}
		case "/ProdCuad":
			val, _ := h.Db.Get(ctx.QueryArgs().Peek("c"))
			if len(val) > 0 {
				var CantProd, w int
				CantProd, w = DecodeSpecialBytes(val[0:2], 200)
			}
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
}
func DecodeSpecialBytes(byte []byte, limit int) (int, int) {
	m := int(byte[0])
	if m < limit {
		return m, 1
	} else {
		return (m-limit)*256 + int(byte[1]) + 200, 2
	}
}
