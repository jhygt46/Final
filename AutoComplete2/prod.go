package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	lediscfg "github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"
	"github.com/valyala/fasthttp"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type MyHandler struct {
	Db           *ledis.DB         `json:"Db"`
	Conf         Config            `json:"Conf"`
	AutoComplete map[string]string `json:"AutoComplete"`
}
type Config struct {
	Tiempo time.Duration `json:"Tiempo"`
}
type Palabras struct {
	Id     uint32 `json:"Id"`
	Tipo   uint32 `json:"Tipo"`
	Nombre string `json:"Nombre"`
}

func main() {

	/*
		lista := make([][]int32, 4)
		lista[0] = make([]int32, 2)
		lista[1] = make([]int32, 2)
		lista[2] = make([]int32, 2)
		lista[3] = make([]int32, 2)

		var i, total int32 = 0, 1114111
		for i = 0; i <= total; i++ {
			x := len([]byte(string(i))) - 1
			if lista[x][0] == 0 {
				lista[x][0] = i
			}
			lista[x][1] = i
		}
		fmt.Println(lista)
		fmt.Println(EncodeAuto())
	*/
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
func EncodeAuto(ListaPalabra []Palabras) string {

	var buf strings.Builder
	for i, pal := range ListaPalabra {
		if i > 0 {
			buf.Write([]byte{44})
		}
		fmt.Fprintf(&buf, "{'I':%v,'N':%v,'T':%v}", pal.Id, pal.Nombre, pal.Tipo)
	}
	return buf.String()
}
func Min2Bytes(bytes []byte) []byte {
	if len(bytes) == 1 {
		b := []byte{0}
		b = append(b, bytes...)
		return b
	}
	return bytes
}
func EncodeSpecialBytes(num int, limit int) ([]byte, bool) {
	max := (255-limit)*256 + 255
	if num <= max {
		if num < limit {
			return []byte{uint8(num)}, true
		} else {
			x := num - limit
			b1 := x/256 + limit
			b2 := x % 256
			return []byte{uint8(b1), uint8(b2)}, true
		}
	} else {
		return nil, false
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
func DecPal(b int) (CantPal int, CantId int, Tipo int) {

	CantPal = b / 12
	aux0 := b % 12
	CantId = aux0 / 4
	aux1 := aux0 % 4
	Tipo = aux1 % 4
	return
}
func (h *MyHandler) SaveDb() {

	lista := make([]int32, 0, 101)
	Count := 0

	for i := 140910; i <= 141010; i++ {
		lista = append(lista, int32(i))
	}

	ListaPalabra := []Palabras{Palabras{Id: 1345, Tipo: 1, Nombre: "CDE"}, Palabras{Id: 1346, Tipo: 1, Nombre: "CDF"}, Palabras{Id: 1347, Tipo: 1, Nombre: "CDE"}, Palabras{Id: 1348, Tipo: 1, Nombre: "CDG"}, Palabras{Id: 1349, Tipo: 1, Nombre: "CDH"}, Palabras{Id: 1350, Tipo: 1, Nombre: "CDK"}, Palabras{Id: 1351, Tipo: 1, Nombre: "CDM"}, Palabras{Id: 1352, Tipo: 1, Nombre: "CDN"}, Palabras{Id: 1353, Tipo: 1, Nombre: "CDO"}}

	bytes := EncodeAuto(ListaPalabra)

	for _, j := range lista {
		for _, k := range lista {
			var b strings.Builder
			fmt.Fprintf(&b, "%c%c", j, k)
			h.AutoComplete[b.String()] = bytes
			Count++
		}
	}
	fmt.Printf("MEMORIA CANTIDAD (%v) LEN (%v) TOTAL (%v)\n", Count, len(string(bytes)), Count*len(string(bytes)))
	Count = 0
	for _, j := range lista {
		for _, k := range lista {
			for _, m := range lista {
				var key = make([]byte, 0)
				key = append(key, Int32to3bytes(j)...)
				key = append(key, Int32to3bytes(k)...)
				key = append(key, Int32to3bytes(m)...)
				h.Db.Set(key, []byte(bytes))
				Count++
			}
		}
	}
	fmt.Printf("DISCO CANTIDAD (%v) LEN (%v) TOTAL (%v)\n", Count, len(string(bytes)), Count*len(string(bytes)))

}
func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {

	if string(ctx.Method()) == "GET" {
		switch string(ctx.Path()) {
		case "/autoCuad":
			/*
				var p []int32
				if err := json.Unmarshal(ctx.QueryArgs().Peek("c"), &p); err == nil {

					var bn []int32
					var key []byte
					var b strings.Builder
					b.Write([]byte{91})

					bn = p[0:2]
					key = GetKey2(bn, ParamBytes(ctx.QueryArgs().Peek("u")))
					val, _ := h.Db.Get(key)
					if len(val) > 0 {
						WriteResponse(&val, 0, &b, p[2:len(p)])
					} else {
						fmt.Println("NOT FOUND DB-CUAD KEY", key)
					}

					b.Write([]byte{93})
					fmt.Fprint(ctx, b.String())

				}
			*/
		case "/auto":

			//now := time.Now()
			var p []int32
			if err := json.Unmarshal(ctx.QueryArgs().Peek("c"), &p); err == nil {

				var leng int = ParamInt(ctx.QueryArgs().Peek("l"))
				var Auto string
				var foundAuto bool
				var bn []int32
				var key []byte
				var Search []byte

				var b strings.Builder
				b.Write([]byte{91})
				for {

					bn = p[0 : len(p)-leng]
					if leng > 0 {
						Search = []byte(string(p[len(p)-leng : len(p)]))
					}

					if Auto, foundAuto = h.AutoComplete[string(bn)]; foundAuto {
						b.WriteString(Auto)
					} else {
						key = GetKey(bn)
						val, _ := h.Db.Get(key)
						if len(val) > 0 {
							WriteResponse(&val, leng, &b, Search)
						} else {
							fmt.Println("DATABASE OUT", bn)
						}
					}
					if leng == 0 {
						break
					}
					leng--
				}
				b.Write([]byte{93})
				fmt.Fprint(ctx, b.String())
			} else {
				fmt.Fprint(ctx, "Error: 15867")
			}
			//fmt.Println("time elapse:", time.Since(now))

		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
}
func WriteResponse(Auto *[]byte, leng int, b *strings.Builder, Search []byte) {

	var i uint8
	var j int = 1
	for i = 0; i < (*Auto)[0]; i++ {
		InfoAuto, w := DecodeSpecialBytes((*Auto)[j:j+2], 200)
		j = j + w
		CantPal, CantId, Tipo := DecPal(InfoAuto)
		for m := 0; m < CantPal; m++ {
			IdPal := GetIntBytesU32((*Auto)[j : j+int(CantId)+2])
			j = j + int(CantId) + 2
			CantNombre := (*Auto)[j]
			Nombre := (*Auto)[j+1 : j+int(CantNombre)+1]
			j = j + int(CantNombre) + 1
			if leng == 0 {
				if (*b).Len() > 1 {
					(*b).Write([]byte{44})
				}
				fmt.Fprintf(b, "{'I':%v,'N':%v,'T':%v}", IdPal, string(Nombre), Tipo)
			} else {
				fmt.Println("BUSCAR MEMORY IdPal:", IdPal, "Nombre:", Nombre, "Tipo", Tipo, "SEARCH:", Search)
			}
		}
	}
}
func ParamBytes(data []byte) []byte {
	var x uint32
	for _, c := range data {
		x = x*10 + uint32(c-'0')
	}
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, x)
	return Reverse(b)
}
func GetIntBytesU32(val []uint8) uint32 {
	switch len(val) {
	case 1:
		return Bytes1toInt32(val[0:1])
	case 2:
		return Bytes2toInt32(val[0:2])
	case 3:
		return Bytes3toInt32(val[0:3])
	case 4:
		return Bytes4toInt32(val[0:4])
	default:
		return 0
	}
}
func Bytes1toInt32(b []uint8) uint32 {
	bytes := make([]byte, 3, 4)
	bytes = append(bytes, b...)
	return binary.BigEndian.Uint32(bytes)
}
func Bytes2toInt32(b []uint8) uint32 {
	bytes := make([]byte, 2, 4)
	bytes = append(bytes, b...)
	return binary.BigEndian.Uint32(bytes)
}
func Bytes3toInt32(b []uint8) uint32 {
	bytes := make([]byte, 1, 4)
	bytes = append(bytes, b...)
	return binary.BigEndian.Uint32(bytes)
}
func Bytes4toInt32(b []uint8) uint32 {
	return binary.BigEndian.Uint32(b)
}
func GetCountBytesInt32(num uint32) int {

	if num <= 255 {
		return 1
	}
	if num <= 65535 {
		return 2
	}
	if num <= 16777215 {
		return 3
	}
	return 4
}
func AddBytes(buf []byte, bytes []byte) []byte {
	return append(buf, bytes...)
}
func IntToBytes(n int) []byte {
	if n == 0 {
		return []byte{0}
	}
	return big.NewInt(int64(n)).Bytes()
}
func GetKey(i []int32) []byte {
	var buf []byte = make([]byte, 0)
	var leng, x int = len(i), 0
	for {

		if i[x] < 256 {
			buf = append(buf, uint8(i[x]))
		}
		if i[x] < 65536 {
			b := make([]byte, 2)
			binary.LittleEndian.PutUint16(b, uint16(i[x]))
			buf = append(buf, Reverse(b)...)
		}
		if i[x] < 16777216 {
			b := make([]byte, 4)
			binary.LittleEndian.PutUint32(b, uint32(i[x]))
			buf = append(buf, Reverse(b[0:3])...)
		}
		x++
		if leng == x {
			break
		}
	}
	return buf
}
func GetKey2(i []int32, cuad []byte) []byte {
	var buf []byte = make([]byte, 0)
	var leng, x int = len(i), 0
	for {

		if i[x] < 256 {
			buf = append(buf, uint8(i[x]))
		}
		if i[x] < 65536 {
			b := make([]byte, 2)
			binary.LittleEndian.PutUint16(b, uint16(i[x]))
			buf = append(buf, Reverse(b)...)
		}
		if i[x] < 16777216 {
			b := make([]byte, 4)
			binary.LittleEndian.PutUint32(b, uint32(i[x]))
			buf = append(buf, Reverse(b[0:3])...)
		}
		x++
		if leng == x {
			break
		}
	}
	buf = append(buf, uint8(0))
	buf = append(buf, cuad...)
	return buf
}
func Inttobytes(i int32) []byte {
	if i < 256 {
		return []byte{uint8(i)}
	} else if i < 65535 {
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(i))
		return Reverse(b)
	} else if i < 16777216 {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(i))
		return Reverse(b[0:3])
	} else {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(i))
		return Reverse(b)
	}
}
func Int32tobytes(i uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return Reverse(b)
}
func Int32to3bytes(i int32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(i))
	return Reverse(b[0:3])
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
func ParamInt32(data []byte) int32 {
	var x int32
	for _, c := range data {
		x = x*10 + int32(c-'0')
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
