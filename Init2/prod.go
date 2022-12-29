// http://localhost:81/init?lang=es&country=43&version=1&upgrade=0
package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	lediscfg "github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"

	"github.com/valyala/fasthttp"
)

type Config struct {
	Tiempo time.Duration `json:"Tiempo"`
}

type MyHandler struct {
	Db           *ledis.DB              `json:"Db"`
	Conf         Config                 `json:"Conf"`
	Language     map[uint8]Lang         `json:"Language"`
	Cuads        map[uint8]Cuads        `json:"Cuads"`
	CuadsB       map[uint8]Cuads        `json:"Cuads"`
	CuadsChilds  map[uint8]CuadsChilds  `json:"CuadsChilds"`
	AutoComplete map[uint8]AutoComplete `json:"AutoComplete"`
	MaxVersion   uint8                  `json:"MaxVersion"`
}
type Lang struct {
	Version map[uint8]*Language `json:"Version"`
}
type Language struct {
	Resp                []byte `json:"Resp"`
	TotalBytes          int    `json:"TotalBytes"`
	UltimaActualizacion uint16 `json:"UltimaActualizacion"`
}
type CuadsChilds struct {
	Lista map[uint32]CuadsList `json:"Lista"`
}
type CuadsList struct {
	Resp                []byte `json:"Resp"`
	TotalBytes          int    `json:"TotalBytes"`
	UltimaActualizacion uint32 `json:"UltimaActualizacion"`
}
type Lista struct {
	Id     uint32   `json:"Id"`
	Points []Points `json:"Points"`
}
type Points struct {
	Lat float64 `json:"Lat"`
	Lng float64 `json:"Lng"`
}
type Cuads struct {
	Resp                []byte `json:"Resp"`
	TotalBytes          int    `json:"TotalBytes"`
	UltimaActualizacion uint16 `json:"UltimaActualizacion"`
}
type AutoComplete struct {
	Auto map[uint8]AutoComp `json:"Auto"`
}
type AutoComp struct {
	Resp                []byte `json:"Resp"`
	TotalBytes          int    `json:"TotalBytes"`
	UltimaActualizacion uint16 `json:"UltimaActualizacion"`
}

func main() {

	cant_paises := 10
	cant_idiomas := 10

	cfg := lediscfg.NewConfigDefault()

	var port string
	if runtime.GOOS == "windows" {
		port = ":81"
		cfg.DataDir = "C:/Go/LedisDB/Init"
	} else {
		port = ":81"
		cfg.DataDir = "/var/Go/LedisDB/Init"
	}

	l, _ := ledis.Open(cfg)
	db, _ := l.Select(0)

	pass := &MyHandler{
		Db:           db,
		Cuads:        make(map[uint8]Cuads, cant_paises),
		CuadsB:       make(map[uint8]Cuads, cant_paises),
		CuadsChilds:  make(map[uint8]CuadsChilds, cant_paises),
		AutoComplete: make(map[uint8]AutoComplete, cant_paises),
		Language:     make(map[uint8]Lang, cant_idiomas),
		Conf:         Config{},
		MaxVersion:   0,
	}

	var zona uint8 = 0
	pass.SaveDb(zona)

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

func (h *MyHandler) SaveDb(zona uint8) {
	/*
		var keyDb []byte

		keyDb = []byte{0, 0}
		listPaises, _ := h.Db.Get(keyDb)
		//h.Db.Set(keyDb, []byte{15, 23, 37, 41, 54, 63, 75, 77, 93, 104, 109, 117, 123, 135, 144, 156, 168})

		keyDb = []byte{0, 1}
		listLanguaje, _ := h.Db.Get(keyDb)
		//h.Db.Set(keyDb, []byte{0, 1, 2, 3})

		for i := 0; i < len(listPaises); i++ {

			// CUADS
			keyCuad := []byte{0, listPaises[i]}

			//h.Db.Set(keyCuad, []byte{0, 0, 91, 123, 34, 73, 34, 58, 49, 44, 34, 80, 34, 58, 91, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 93, 125, 44, 123, 34, 73, 34, 58, 50, 44, 34, 80, 34, 58, 91, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 93, 125, 44, 123, 34, 73, 34, 58, 51, 44, 34, 80, 34, 58, 91, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 93, 125, 44, 123, 34, 73, 34, 58, 52, 44, 34, 80, 34, 58, 91, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 93, 125, 44, 123, 34, 73, 34, 58, 53, 44, 34, 80, 34, 58, 91, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 44, 123, 34, 65, 34, 58, 51, 52, 46, 53, 52, 53, 54, 44, 34, 78, 34, 58, 49, 57, 48, 46, 53, 51, 51, 50, 125, 93, 125, 93})

			verCuads, _ := h.Db.Get(keyCuad)
			cuad := Cuads{UltimaActualizacion: ParamInt16(verCuads[0:2]), Resp: verCuads[2:]}
			cuad.TotalBytes = len(cuad.Resp)
			h.Cuads[uint8(listPaises[i])] = cuad

			// AUTOCOMPLETE
			keyAuto := []byte{2, listPaises[i]}
			listaAuto, _ := h.Db.Get(keyAuto)

			h.AutoComplete[uint8(listPaises[i])] = AutoComplete{Auto: make(map[uint8]AutoComp, 0)}

			for j := 0; j < len(listaAuto); j++ {

				keyAuto2 := []byte{2, listPaises[i], listaAuto[j]}

				//h.Db.Set(keyAuto2, []byte{123, 34, 65, 34, 58, 91, 123, 34, 73, 34, 58, 49, 44, 34, 78, 34, 58, 34, 109, 98, 117, 108, 97, 110, 99, 105, 97, 34, 44, 34, 84, 34, 58, 49, 125, 44, 123, 34, 73, 34, 58, 50, 44, 34, 78, 34, 58, 34, 118, 105, 111, 110, 101, 115, 34, 125, 93, 44, 34, 66, 34, 58, 91, 123, 34, 73, 34, 58, 51, 44, 34, 78, 34, 58, 34, 97, 114, 101, 115, 34, 125, 44, 123, 34, 73, 34, 58, 52, 44, 34, 78, 34, 58, 34, 111, 98, 101, 100, 97, 115, 34, 125, 93, 44, 34, 67, 34, 58, 91, 123, 34, 73, 34, 58, 53, 44, 34, 78, 34, 58, 34, 111, 110, 99, 105, 101, 114, 116, 111, 115, 34, 125, 44, 123, 34, 73, 34, 58, 54, 44, 34, 78, 34, 58, 34, 105, 110, 101, 115, 34, 125, 93, 44, 34, 68, 34, 58, 91, 123, 34, 73, 34, 58, 55, 44, 34, 78, 34, 58, 34, 97, 100, 111, 115, 34, 125, 44, 123, 34, 73, 34, 58, 56, 44, 34, 78, 34, 58, 34, 111, 114, 105, 116, 111, 115, 34, 125, 93, 125})

				verAuto, _ := h.Db.Get(keyAuto2)
				auto := AutoComp{UltimaActualizacion: ParamInt16(verAuto[0:2]), Resp: verAuto[2:]}
				auto.TotalBytes = len(auto.Resp)
				h.AutoComplete[uint8(listPaises[i])].Auto[listaAuto[j]] = auto

			}

			// CUADSCHILDS
			keyCuadChilds := []byte{3, listPaises[i]}
			listaCuadChilds, _ := h.Db.Get(keyCuadChilds)

			h.CuadsChilds[uint8(listPaises[i])] = CuadsChilds{Lista: make(map[uint32]CuadsList, 0)}

			for j := 0; j < len(listaCuadChilds); j = j + 4 {

				keyCC := append([]byte{3}, listaCuadChilds[j:j+4]...)
				verCuadChilds, _ := h.Db.Get(keyCC)

				childCuads := CuadsList{UltimaActualizacion: GetIntBytesU32(verCuadChilds[0:2]), Resp: verCuadChilds[2:]}
				childCuads.TotalBytes = PrintCuadsBytesLen(childCuads.Resp)
				h.CuadsChilds[uint8(listPaises[i])].Lista[binary.BigEndian.Uint32(listaCuadChilds[j:j+4])] = childCuads

			}
		}

		// LANGUAJE
		for j := 0; j < len(listLanguaje); j++ {

			keyLang1 := []byte{1, listLanguaje[j], 0}
			verLang1, _ := h.Db.Get(keyLang1)

			keyLang2 := []byte{1, listLanguaje[j], 1}
			verLang2, _ := h.Db.Get(keyLang2)

			//h.Db.Set(keyLang, []byte{0, 0, 91, 91, 34, 72, 79, 76, 65, 45, 65, 49, 34, 44, 34, 72, 79, 76, 65, 45, 65, 50, 34, 44, 34, 72, 79, 76, 65, 45, 65, 51, 34, 93, 44, 91, 34, 72, 79, 76, 65, 45, 66, 49, 34, 44, 34, 72, 79, 76, 65, 45, 66, 50, 34, 44, 34, 72, 79, 76, 65, 45, 66, 51, 34, 93, 44, 91, 34, 72, 79, 76, 65, 45, 67, 49, 34, 44, 34, 72, 79, 76, 65, 45, 67, 50, 34, 44, 34, 72, 79, 76, 65, 45, 67, 51, 34, 93, 93})

			lang := Language{UltimaActualizacion: ParamInt16(verLang1[0:2]), Resp: verLang1[2:]}
			lang.TotalBytes = len(lang.Resp)

			for z := 0; z < len(verLang2); z = z + 2 {

				//l := Lang{Version: make(map[uint8]*Language, 0)}
				//l.Version[verLang2[z+1:z+2]] = &lang
				//h.Language[verLang2[z:z+1]] = l

			}
		}
	*/
	ListaCuads := []Lista{
		Lista{
			Id: 1,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 2,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 3,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 4,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 5,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 6,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 7,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 8,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 9,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 10,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 11,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 12,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 13,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 14,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 15,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
		Lista{
			Id: 16,
			Points: []Points{
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
				Points{Lat: -12.5345765, Lng: -88.5345765},
			},
		},
	}

	ecc := EncodeCuadsChilds(ListaCuads)
	x := PrintCuadsSinLen(ecc)

	cuad1 := Cuads{UltimaActualizacion: 0, Resp: x}
	h.Cuads[0] = cuad1

	cuad2 := Cuads{UltimaActualizacion: 0, Resp: ecc, TotalBytes: len(x)}
	h.CuadsB[0] = cuad2

	fmt.Println("LEN A: ", len(x))
	fmt.Println("LEN B: ", len(ecc))

	//key := append([]byte{0}, Int32tobytes(id)...)
	//h.Db.Set(key, buf)

}

func (h *MyHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {

	if string(ctx.Method()) == "GET" {
		switch string(ctx.Path()) {
		case "/init":
			/*
				ctx.Response.Header.Set("Content-Type", "application/json")

				lan := ParamInt(ctx.QueryArgs().Peek("lang"))
				country := ParamInt(ctx.QueryArgs().Peek("country"))
				version := ParamInt(ctx.QueryArgs().Peek("version"))
				upgrade := ParamInt16(ctx.QueryArgs().Peek("upgrade"))

				var grow int = 2
				var setGrow bool = true
				var coma bool = false

				var b1 []byte
				var b2 []byte
				var b3 []byte

				var auxb1 []byte
				var auxb2 []byte
				var auxb3 []byte
				var auxb4 []byte
				var auxb5 []byte

				auxb1 = []byte{123}

				if Lang, Found := h.Language[country]; Found {
					if upgrade <= *&Lang.Version[version].UltimaActualizacion {
						if version <= h.MaxVersion {
							auxb2 = []byte{34, 76, 34, 58}
							grow = grow + *&Lang.Version[version].TotalBytes + 4
							b1 = *&Lang.Version[version].Resp
							coma = true
						}
					}
				}
				if Cuads, Found := h.Cuads[country]; Found {
					if upgrade <= Cuads.UltimaActualizacion {
						if coma {
							auxb3 = []byte{44, 34, 67, 34, 58}
							grow = grow + Cuads.TotalBytes + 5
						} else {
							auxb3 = []byte{34, 67, 34, 58}
							grow = grow + Cuads.TotalBytes + 4
						}
						b2 = Cuads.Resp
						coma = true
					}
				}
				if Auto, Found := h.AutoComplete[country]; Found {
					if Resp, Found := Auto.Auto[lan]; Found {
						if coma {
							auxb4 = []byte{44, 34, 65, 34, 58}
							grow = grow + Resp.TotalBytes + 5
						} else {
							auxb4 = []byte{34, 65, 34, 58}
							grow = grow + Resp.TotalBytes + 4
						}
						b3 = Resp.Resp
					}
				}

				auxb5 = []byte{125}

				var b strings.Builder
				if setGrow {
					b.Grow(grow)
				}

				b.Write(auxb1)
				b.Write(auxb2)
				b.Write(b1)
				b.Write(auxb3)
				b.Write(b2)
				b.Write(auxb4)
				b.Write(b3)
				b.Write(auxb5)

				fmt.Fprint(ctx, b.String())
			*/
		case "/cuads":
			/*
				ctx.Response.Header.Set("Content-Type", "application/json")

				country := ParamInt(ctx.QueryArgs().Peek("country"))
				id := ParamInt32(ctx.QueryArgs().Peek("id"))
				upgrade := ParamInt32(ctx.QueryArgs().Peek("upgrade"))

				if Childs, Found := h.CuadsChilds[country]; Found {
					if Lista, Found := Childs.Lista[id]; Found {
						if Lista.UltimaActualizacion <= upgrade {
							ctx.SetBody(PrintCuadsBytes2(Lista.TotalBytes, Lista.Resp))
						}
					} else {
						key := ctx.QueryArgs().Peek("id")
						val, _ := h.Db.Get(key)
						if len(val) > 0 {
							if Bytes2toInt32(val[0:2]) <= upgrade {
								ctx.SetBody(PrintCuadsBytes2(int(binary.BigEndian.Uint32(val[2:6])), val[6:]))
							}
						}
					}
				}
			*/
		case "/cuadsA":

			ctx.Response.Header.Set("Content-Type", "application/json")

			if Cuad, Found := h.Cuads[0]; Found {
				if Cuad.UltimaActualizacion <= 0 {
					ctx.SetBody(Cuad.Resp)
				}
			}

		case "/cuadsB":

			ctx.Response.Header.Set("Content-Type", "application/json")

			if Cuad, Found := h.CuadsB[0]; Found {
				if Cuad.UltimaActualizacion <= 0 {
					ctx.SetBody(PrintCuadsConLen(Cuad.TotalBytes, Cuad.Resp))
				}
			}

		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
}

func PrintCuadsBytesLen(lista []byte) int {
	//PrintCuadsBytes2(&b, lista)
	return 0
}
func EncodeCuadsChilds(lista []Lista) []byte {

	var buf []byte
	buf = AddBytes(buf, Int32tobytes(uint32(len(lista))))
	for _, x := range lista {
		buf = AddBytes(buf, Int32tobytes(x.Id))
		buf = AddBytes(buf, Int32tobytes(uint32(len(x.Points))))
		for _, y := range x.Points {
			buf = AddBytes(buf, newFloat32tobytes(y.Lat))
			buf = AddBytes(buf, newFloat32tobytes(y.Lng))
		}
	}
	return buf
}

/*
func PrintCuadsBytes(b *strings.Builder, lista []byte) {

	var i uint32
	var j uint32
	var id uint32
	var lat float32
	var lng float32
	var CantPoints uint32
	CantCuads := Bytes4toInt32(lista[0:4])
	var x int = 4

	(*b).Write([]byte{91})

	for i = 0; i < CantCuads; i++ {
		id = Bytes4toInt32(lista[x : x+4])
		CantPoints = Bytes4toInt32(lista[x+4 : x+8])
		x = x + 8

		if i > 0 {
			(*b).Write([]byte{44})
		}

		(*b).Write([]byte{123, 34, 73, 34, 58})
		fmt.Fprintf(b, "%v", id)
		(*b).Write([]byte{44, 34, 80, 34, 58, 91})

		for j = 0; j < CantPoints; j++ {
			lat = Float32frombytes(lista[x : x+4])
			lng = Float32frombytes(lista[x+4 : x+8])
			x = x + 8
			if j > 0 {
				(*b).Write([]byte{44})
			}
			fmt.Fprintf(b, "{'A':%v,'N':%v}", lat, lng)
		}

		(*b).Write([]byte{93})

	}

	(*b).Write([]byte{93})
}
*/
func PrintCuadsConLen(len int, lista []byte) []byte {

	var x, y int = 4, 0
	var lat, lng float64
	var i, j, id, CantPoints uint32
	var res []byte = make([]byte, len)
	var CantCuads uint32 = binary.BigEndian.Uint32(lista[0:4])

	for i = 0; i < CantCuads; i++ {

		id = binary.BigEndian.Uint32(lista[x : x+4])
		CantPoints = binary.BigEndian.Uint32(lista[x+4 : x+8])
		x = x + 8

		if i > 0 {
			res[y] = 44
			y++
		}

		y += copy(res[y:], []byte{91, 123, 34, 73, 34, 58})
		y += copy(res[y:], PrintintBytes(id))
		y += copy(res[y:], []byte{44, 34, 80, 34, 58, 91})

		for j = 0; j < CantPoints; j++ {

			lat = newFloat32frombytes(lista[x : x+4])
			lng = newFloat32frombytes(lista[x+4 : x+8])
			x = x + 8

			if j > 0 {
				res[y] = 44
				y++
			}

			y += copy(res[y:], []byte{123, 34, 65, 34, 58})
			y += copy(res[y:], PrintfloatBytes(lat))
			y += copy(res[y:], []byte{44, 34, 78, 34, 58})
			y += copy(res[y:], PrintfloatBytes(lng))
			res[y] = 125
			y++

		}

		res[y] = 93
		y++

	}

	res[y] = 93
	return res
}
func PrintCuadsSinLen(lista []byte) []byte {

	var x int = 4
	var lat, lng float64
	var i, j, id, CantPoints uint32
	var res []byte
	var CantCuads uint32 = binary.BigEndian.Uint32(lista[0:4])

	for i = 0; i < CantCuads; i++ {

		id = binary.BigEndian.Uint32(lista[x : x+4])
		CantPoints = binary.BigEndian.Uint32(lista[x+4 : x+8])
		x = x + 8

		if i > 0 {
			res = AddBytes(res, []byte{44})
		}

		res = AddBytes(res, []byte{91, 123, 34, 73, 34, 58})
		res = AddBytes(res, PrintintBytes(id))
		res = AddBytes(res, []byte{44, 34, 80, 34, 58, 91})

		for j = 0; j < CantPoints; j++ {

			lat = newFloat32frombytes(lista[x : x+4])
			lng = newFloat32frombytes(lista[x+4 : x+8])
			x = x + 8

			if j > 0 {
				res = AddBytes(res, []byte{44})
			}

			res = AddBytes(res, []byte{123, 34, 65, 34, 58})
			res = AddBytes(res, PrintfloatBytes(lat))
			res = AddBytes(res, []byte{44, 34, 78, 34, 58})
			res = AddBytes(res, PrintfloatBytes(lng))
			res = AddBytes(res, []byte{125})

		}

		res = AddBytes(res, []byte{93})

	}

	res = AddBytes(res, []byte{93})
	return res
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
func Int32tobytes(i uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return Reverse(b)
}
func ParamIntBase(data []byte) int {
	var x int
	for _, c := range data {
		x = x*10 + int(c-'0')
	}
	return x
}
func Reverse(numbers []uint8) []byte {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}
func Int16toByte(i uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return Reverse(b)
}
func ParamInt(data []byte) uint8 {
	var x uint8
	for _, c := range data {
		x = x*10 + uint8(c-'0')
	}
	return x
}
func ParamInt16(data []byte) uint16 {
	var x uint16
	for _, c := range data {
		x = x*10 + uint16(c-'0')
	}
	return x
}
func ParamInt32(data []byte) uint32 {
	var x uint32
	for _, c := range data {
		x = x*10 + uint32(c-'0')
	}
	return x
}
func Float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}
func float32ToByte(f float32) []byte {
	bits := math.Float32bits(f)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}
func GetCountBytesInt(num int) int {

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
func PrintCopyPasteBytes(s string) {

	a := []byte(s)
	fmt.Printf("{")
	for i, b := range a {
		if i > 0 {
			fmt.Printf(",")
		}
		if b == 39 {
			a[i] = 34
		}
		fmt.Printf("%v", a[i])
	}
	fmt.Printf("}")
}
func PointInPolygon(lat float64, lng float64, poly []Points) bool {

	var j int = 0
	var nverts int = len(poly)
	var in bool = false
	var verts []Points = poly

	for i := 1; i < nverts; i++ {
		if ((verts[i].Lng > lng) != (verts[j].Lng > lng)) &&
			(lat < (verts[j].Lat-verts[i].Lat)*(lng-verts[i].Lng)/(verts[j].Lng-verts[i].Lng)+verts[i].Lat) {
			in = !in
		}
		j = i
	}
	return in
}
func PrintintBytes(x uint32) []byte {
	var i uint32 = 0
	res := make([]byte, 10)
	for {
		if x > 0 {
			res[9-i] = uint8(x%10 + 48)
		} else {
			break
		}
		x = x / 10
		i++
	}
	return res[10-i:]
}
func PrintfloatBytes(x float64) []byte {

	res := make([]byte, 12)
	var aux1, aux2 float64
	var i, j int = 0, 0

	if x < 0 {
		res[0] = 45
		aux1 = -float64(int(x))
		aux2 = (x + aux1) * -10000000
		i = 1
	} else {
		aux1 = float64(int(x))
		aux2 = (x - aux1) * 10000000
	}

	b1, n1 := PrintintBytes2(int(aux1))
	b2, _ := PrintintBytes2(int(aux2))

	copy(res[i:], b1)
	i = i + n1
	res[i] = 46
	copy(res[i+1:], b2)
	for j = 11; j >= 0; j-- {
		if res[j] != 0 && res[j] != 48 {
			break
		}
	}
	return res[0 : j+1]
}
func PrintintBytes2(x int) ([]byte, int) {
	var i int = 0
	res := make([]byte, 10)
	for {
		if x > 0 {
			res[9-i] = uint8(x%10 + 48)
		} else {
			break
		}
		x = x / 10
		i++
	}
	return res[10-i:], i
}
func newFloat32frombytes(b []byte) float64 {
	var x int = int(b[3]) + int(b[2])*256 + int(b[1])*65536 + int(b[0])*16777216 - 1800000000
	return float64(x) / 10000000
}
func newFloat32tobytes(f float64) []byte {
	b := make([]byte, 4)
	var x int = int(f*10000000 + 1800000000)
	var r int = x % 16777216
	b[0] = uint8(x / 16777216)
	b[1] = uint8(r / 65536)
	r = r % 65536
	b[2], b[3] = uint8(r/256), uint8(r%256)
	return b
}

func GetIntBytesU32(val []uint8) uint32 {
	var len = len(val)
	if len == 4 {
		return binary.BigEndian.Uint32(val)
	} else if len == 3 {
		return Bytes3toInt32(val)
	} else if len == 2 {
		return Bytes2toInt32(val)
	} else {
		return uint32(val[0])
	}
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
