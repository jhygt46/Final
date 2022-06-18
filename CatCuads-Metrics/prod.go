package main

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"

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
	l, _ := ledis.Open(cfg)
	db, _ := l.Select(0)

	h := &MyHandler{
		Db:    db,
		Cuads: make(map[uint32][]byte, 0),
	}

	h.Cuads[1] = []byte{255, 255, 255, 255, 255, 255, 255, 255}

	var port string
	if runtime.GOOS == "windows" {
		port = ":81"
	} else {
		port = ":80"
	}

	fasthttp.ListenAndServe(port, h.HandleFastHTTP)

}

func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {

	if string(ctx.Method()) == "GET" {
		switch string(ctx.Path()) {
		case "/":
			now := time.Now()

			var p Params
			if err := json.Unmarshal(ctx.QueryArgs().Peek("p"), &p); err == nil {

				c := ParamUint32(ctx.QueryArgs().Peek("c"))
				if cuad, foundCuad := h.Cuads[c]; foundCuad {

					var b strings.Builder
					var first bool = true
					var num uint32
					var bytes uint32
					var posbyte uint32
					var arrcuad uint32

					b.Grow(50)
					b.Write([]byte{91})

					for _, arrcuad = range p.C {

						num = arrcuad - 1
						bytes = num / 8
						posbyte = num % 8

						if DecBits(cuad[bytes], int(posbyte)) == 0 {
							if first {
								first = false
								fmt.Fprintf(&b, "'%v'", arrcuad)
							} else {
								fmt.Fprintf(&b, ",'%v'", arrcuad)
							}
						}

					}

					b.Write([]byte{93})
					fmt.Fprint(ctx, b.String())

				} else {
					fmt.Fprintf(ctx, "ErrorFoundCuads")
				}

			} else {
				fmt.Fprintf(ctx, "ErrorDecode")
			}
			fmt.Println("time elapse:", time.Since(now))

		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
}

func (h *MyHandler) AddCuad(cat uint32, arrcuad uint32) {

	var num uint32
	var bytes uint32
	var posbyte uint32

	num = arrcuad - 1
	bytes = num / 8
	posbyte = num % 8

	if cuad, foundCuad := h.Cuads[cat]; foundCuad {
		if DecBits(cuad[bytes], 7-int(posbyte)) == 0 {
			h.Cuads[cat][bytes] = h.Cuads[cat][bytes] + uint8(math.Pow(2, float64(posbyte)))
		}
	}
}
func (h *MyHandler) RmCuad(cat uint32, arrcuad uint32) {

	var num uint32
	var bytes uint32
	var posbyte uint32

	num = arrcuad - 1
	bytes = num / 8
	posbyte = num % 8

	if cuad, foundCuad := h.Cuads[cat]; foundCuad {
		if DecBits(cuad[bytes], 7-int(posbyte)) == 1 {
			h.Cuads[cat][bytes] = h.Cuads[cat][bytes] - uint8(math.Pow(2, float64(posbyte)))
		}
	}
}
func ParamUint32(data []byte) uint32 {
	var x uint32
	for _, c := range data {
		x = x*10 + uint32(c-'0')
	}
	return x
}
func DecBits(b byte, n int) uint8 {

	var value uint8 = 0
	var aux uint8 = b
	var div uint8 = 128
	for i := 0; i < n+1; i++ {
		value = aux / div
		aux = aux % div
		div = div / 2
		if i == n {
			break
		}
	}
	return value
}
