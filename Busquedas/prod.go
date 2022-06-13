package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
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

type Respuesta struct {
	Prods []ResProd `json:"Prods"`
	Emps  []ResEmp  `json:"Emps"`
	Count int       `json:"Count"`
}
type ResProd struct {
	Id        uint32  `json:"Id"`
	Nombre    string  `json:"Nombre"`
	Precio    uint32  `json:"Precio"`
	Calidad   uint32  `json:"Calidad"`
	Distancia uint32  `json:"Distancia"`
	Nota      float64 `json:"Nota"`
	TipoId    uint8   `json:"TipoId"`
	IdEmp     uint32  `json:"IdEmp"`
}
type ResEmp struct {
	Id     uint32  `json:"Id"`
	Nombre string  `json:"Nombre"` // Nombre
	Lat    float32 `json:"Lat"`    // Lat
	Lng    float32 `json:"Lng"`    // Lng
	Count  float32 `json:"Count"`  // Count
}
type MyHandler struct {
	Db       *ledis.DB         `json:"Db"`
	Prods    map[uint32][]byte `json:"Prods"`
	Empresas map[uint32][]byte `json:"Empresas"`
	Catalogo map[uint32][]byte `json:"Catalogo"`
	Conf     Config            `json:"Conf"`
}
type Config struct {
	Tiempo time.Duration `json:"Tiempo"`
}
type Empresa struct {
	Id     uint32  `json:"Id"`
	IdLoc  uint32  `json:"IdLoc"`
	IdCat  uint32  `json:"IdCat"`
	Lat    float32 `json:"Lat"`
	Lng    float32 `json:"Lng"`
	Nombre string  `json:"Nombre"`
	Prods  []Prods `json:"Prods"`
}
type Prods struct {
	Id       uint64    `json:"Id"`
	Tipo     int       `json:"Tipo"` // 0 ID - 1 IDS
	Precio   uint64    `json:"Precio"`
	Nombre   string    `json:"Nombre"`
	Calidad  uint8     `json:"Calidad"`
	Cantidad uint8     `json:"Cantidad"`
	Filtros  []Filtros `json:"Filtros"`
	Evals    []uint8   `json:"Opciones"`
}
type Filtros struct {
	N uint8              `json:"N"`
	T uint8              `json:"T"`
	V []uint16           `json:"V"`
	R []uint32           `json:"R"`
	P []FiltroconPrecios `json:"P"`
}
type FiltroconPrecios struct {
	Id     uint16 `json:"Id"`
	Precio uint32 `json:"Precio"`
}
type Params struct {
	C []uint32   `json:"C"` // CUADRANTES
	F [][]uint32 `json:"F"` // FILTROS
	E []uint8    `json:"E"` // EVALS
	D uint16     `json:"D"` // DESDE
	N uint8      `json:"N"` // NUMERO DE FILTROS
	L int        `json:"L"` // LARGO
	O []float64  `json:"O"` // OPCIONES 1 PRECIO - 2 DISTANCIA - 3 CALIDAD
}
type NewParams struct {
	F    [][]uint32 `json:"F"`    // FILTROS
	E    []uint8    `json:"E"`    // EVALS
	D    uint16     `json:"D"`    // DESDE
	O    []float64  `json:"O"`    // OPCIONES 1 PRECIO - 2 DISTANCIA - 3 CALIDAD
	L    int        `json:"L"`    // LARGO
	SO   float64    `json:"SO"`   // SUMAOPCIONES
	SE   uint32     `json:"SE"`   // SUMAEVAL
	Elen int        `json:"Elen"` // SUMAEVAL
	Flen int        `json:"Flen"` // SUMAEVAL
}

func main() {

	cfg := lediscfg.NewConfigDefault()

	var port string
	if runtime.GOOS == "windows" {
		port = ":81"
		cfg.DataDir = "C:/Go/Pruebas/LedisDb/var"
	} else {
		port = ":80"
		cfg.DataDir = "C:/Go/Pruebas/LedisDb/var"
	}

	l, _ := ledis.Open(cfg)
	db, _ := l.Select(0)

	pass := &MyHandler{
		Db:       db,
		Prods:    make(map[uint32][]byte, 0),
		Empresas: make(map[uint32][]byte, 0),
		Catalogo: make(map[uint32][]byte, 0),
	}

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
			now := time.Now()
			var p Params
			if err := json.Unmarshal(ctx.QueryArgs().Peek("p"), &p); err == nil {
				if len(p.O) == 4 {
					var key []byte
					P := NewParams{D: p.D}
					if SO := p.O[0] + p.O[1] + p.O[2]; SO > 0 {
						P.O = p.O
						P.SO = SO
					}
					if Flen := len(p.F); Flen > 0 {
						P.F = p.F
						P.Flen = Flen
					}
					if Elen := len(p.E); Elen > 0 {
						P.E = p.E
						P.Elen = Elen
						var sumE uint32 = 0
						for _, x := range p.E {
							sumE = sumE + uint32(x)
						}
						if sumE > 0 {
							P.E = p.E
							P.Elen = Elen
							P.SE = sumE
						}
					}
					if p.L < 19 || p.L > 51 {
						P.L = 20
					}
					Res := Respuesta{Prods: make([]ResProd, 0, P.L), Emps: make([]ResEmp, 0, P.L), Count: 0}
					cat := ParamBytes(ctx.QueryArgs().Peek("c"))
					for _, cuad := range p.C {
						key = append(cat, Int32tobytes(cuad)...)
						val, _ := h.Db.Get(key)
						h.DecodeBytes(&Res, val, P)
					}
					var b strings.Builder
					b.Grow(900)
					for m, p := range Res.Prods {
						if m == 0 {
							b.Write([]byte{123, 39, 80, 39, 58, 91})
							fmt.Fprintf(&b, "{'Id':%d,'Nombre':'%s','Precio':%v,'Nota':%v}", p.Id, p.Nombre, p.Precio, p.Nota)
						} else {
							fmt.Fprintf(&b, ",{'Id':%d,'Nombre':'%s','Precio':%v,'Nota':%v}", p.Id, p.Nombre, p.Precio, p.Nota)
						}
					}
					for n, e := range Res.Emps {
						if n == 0 {
							b.Write([]byte{93, 44, 39, 69, 39, 58, 91})
							fmt.Fprintf(&b, "{'Id':%d,'Nombre':'%s'}", n, e.Nombre)
						} else {
							fmt.Fprintf(&b, "{'Id':%d,'Nombre':'%s'}", n, e.Nombre)
						}
					}
					b.Write([]byte{93, 125})
					fmt.Fprint(ctx, b.String())
				} else {
					fmt.Fprintf(ctx, "ErrorHTTP")
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
func (h *MyHandler) DecodeBytes(Res *Respuesta, bytes []byte, P NewParams) {

	var length int = len(bytes)
	var IdEmp uint32
	var IdProd uint32
	var Precio uint32
	var Calidad uint32
	var NombreEmp string
	var NElen int
	var Size uint32
	var Lat float32
	var Lng float32
	var Value []byte
	var j int = 0
	var x int = 0
	var Distancia uint32 = 0
	var Nota float64 = 0
	var NotaMenor float64 = 999999
	var Posicion int = 0
	var Id_Emp uint32 = 0
	var CountFiltro float64 = 0

	var x1 float64 = 1800
	var y1 float64 = 5000
	var m float64 = -0.02523

	for {

		Size = 0
		EmpCache, CatalogoCache, EmpSize, Loc, ByteSize := DecEmp(bytes[j])
		IdEmp = GetIntBytesU32(bytes[j+1 : j+3+int(EmpSize)])
		j = j + int(EmpSize) + 3

		if EmpCache == 1 {
			if Emp, foundEmp := h.Empresas[IdEmp]; foundEmp {

				NElen = int(Emp[0]) + 1
				NombreEmp = string(Emp[1:NElen])
				IdLoc := GetIntBytesU32(bytes[j : j+int(Loc)+1])
				Pos := NElen + int(IdLoc)*8
				Lat = Float32frombytes(Emp[Pos : Pos+4])
				Lng = Float32frombytes(Emp[Pos+4 : Pos+8])

				j = j + int(Loc) + 1

				if CatalogoCache == 1 {
					IdCat := GetIntBytesU32(bytes[j : j+int(ByteSize)+1])
					j = j + int(ByteSize) + 1
					if Cat, foundCat := h.Catalogo[IdCat]; foundCat {
						Value = Cat
					}
				} else {
					Size = GetIntBytesU32(bytes[j : j+int(ByteSize)+1])
					j = j + 1 + int(ByteSize)
					Value = bytes[j : j+int(Size)]
				}
			}
		} else {

			Lat = Float32frombytes(bytes[j : j+4])
			Lng = Float32frombytes(bytes[j+4 : j+8])
			j = j + 8
			NElen = int(bytes[j])
			NombreEmp = string(bytes[j+1 : j+1+NElen])
			j = j + 1 + NElen
			Size = GetIntBytesU32(bytes[j : j+int(ByteSize)+1])
			j = j + 1 + int(ByteSize)
			Value = bytes[j : j+int(Size)]
		}

		if Distancia = Distance(-33.44546, 70.44546, Lat, Lng); Distancia > 0 {

			x = 0
			cantarr := Value[x]
			x++
			for i := uint8(0); i < cantarr; i++ {

				_, CantPrecio, CantId, Eval, Filtro, TipoId := DecProd(Value[x])
				CantProd, w := DecodeSpecialBytes(Value[x+1:x+3], 200)
				x = x + w + 1

				for s := 0; s < CantProd; s++ {

					Res.Count++

					IdProd = GetIntBytesU32(Value[x : x+int(CantId)+2])
					x = x + int(CantId) + 2

					Precio = GetIntBytesU32(Value[x : x+int(CantPrecio)+2])
					x = x + int(CantPrecio) + 2

					if TipoId == 1 {

						nlen := int(Value[x])
						Nombre := string(Value[x+1 : x+1+nlen])
						x = x + 1 + nlen
						Calidad = uint32(Value[x])

						x++
						if Filtro == 1 {
							CantFiltro, t := DecodeSpecialBytes(Value[x:x+2], 200)
							if CantFiltro > 0 {
								if P.Flen > 0 {
									CountFiltro = CompareFiltro(Value[x+t:x+t+int(CantFiltro)], P.F)
								}
								x = x + t + int(CantFiltro)
							} else {
								x = x + 1
							}
						}
						if Eval == 1 {
							CantEval := Value[x]
							if P.Elen > 0 {
								Calidad = CompareEval(Value[x+1:x+1+int(CantEval)], P.E, P.SE)
								//fmt.Println("Calidad:", Calidad)
							}
							x = x + 1 + int(CantEval)
						}

						Nota = (m*(float64(Precio)-x1)+y1)*(P.O[0]/P.SO) + (m*(float64(Distancia)-x1)+y1)*(P.O[1]/P.SO) + (m*(float64(Calidad)-x1)+y1)*(P.O[2]/P.SO) + CountFiltro*1000

						if len(Res.Prods) < P.L {

							Res.Prods = append(Res.Prods, ResProd{Id: IdProd, Distancia: Distancia, Nombre: Nombre, Precio: Precio, Calidad: Calidad, Nota: Nota, IdEmp: IdEmp, TipoId: TipoId})
							var b bool = true
							for q, emp := range Res.Emps {
								if emp.Id == IdEmp {
									b = false
									Res.Emps[q].Count = emp.Count + 1
								}
							}
							if b {
								Res.Emps = append(Res.Emps, ResEmp{Id: IdEmp, Nombre: NombreEmp, Count: 1, Lat: Lat, Lng: Lng})
							}
							if NotaMenor > Nota {
								NotaMenor = Nota
							}

						} else {

							if Nota > NotaMenor {

								NotaMenor, Posicion, Id_Emp = GetNotaMenor(Res.Prods, Nota)
								if Posicion > -1 {

									var b bool = true
									var vacio int = -1
									for r, emp := range Res.Emps {
										if emp.Id == Id_Emp {
											if emp.Count == 1 {
												Res.Emps[r].Id = 0
												Res.Emps[r].Count = 0
												vacio = r
											}
											if emp.Count > 1 {
												Res.Emps[r].Count = emp.Count - 1
											}
										}
										if emp.Id == IdEmp {
											b = false
											Res.Emps[r].Count = emp.Count + 1
										}
										if emp.Id == 0 {
											vacio = r
										}
									}
									if b {
										if vacio == -1 {
											Res.Emps = append(Res.Emps, ResEmp{Id: IdEmp, Nombre: NombreEmp, Count: 1, Lat: Lat, Lng: Lng})
										} else {
											Res.Emps[vacio] = ResEmp{Id: IdEmp, Nombre: NombreEmp, Count: 1, Lat: Lat, Lng: Lng}
										}
									}

									Res.Prods[Posicion] = ResProd{Id: IdProd, Distancia: Distancia, Nombre: Nombre, Precio: Precio, Calidad: Calidad, Nota: Nota, IdEmp: IdEmp, TipoId: TipoId}

								}
							}

						}
					} else {

						if prod, found := h.Prods[IdProd]; found {

							CantString, ProF, ProE := DecProMem(prod[0])

							nlen := int(CantString)
							NombreP := string(prod[1 : 1+nlen])
							Calidad = uint32(prod[1+nlen])
							d := 2 + nlen

							Silence(NombreP, d)

							Nota = (m*(float64(Precio)-x1)+y1)*(P.O[0]/P.O[3]) + (m*(float64(Distancia)-x1)+y1)*(P.O[1]/P.O[3]) + (m*(float64(Calidad)-x1)+y1)*(P.O[2]/P.O[3])

							if len(Res.Prods) < P.L {
								Res.Prods = append(Res.Prods, ResProd{Id: IdProd, Distancia: Distancia, Nombre: NombreP, Precio: Precio, Calidad: Calidad, Nota: Nota, IdEmp: IdEmp, TipoId: TipoId})
							} else {

							}

							if ProF == 1 {
								//fmt.Println("CACHE FILTRO")
							}
							if ProE == 1 {
								//fmt.Println("CACHE EVALS")
							}

							//fmt.Println(prod, Filtro, Eval)
							//fmt.Printf("SICACHE IdProd %v - Precio %v - Nombre %s - Calidad %v\n", IdProd, Precio, Nombre, Calidad)
						} else {
							//fmt.Println("ERROR PROD NOT FOUND", IdProd)
						}
					}
				}
			}

			j = j + int(Size)
		} else {
			j = j + int(Size)
		}

		if length <= j {
			break
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
func Reverse(numbers []uint8) []uint8 {
	for i := 0; i < len(numbers)/2; i++ {
		j := len(numbers) - i - 1
		numbers[i], numbers[j] = numbers[j], numbers[i]
	}
	return numbers
}
func Int32tobytes(i uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
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
func Bytes2toInt16(b []uint8) uint16 {
	return binary.BigEndian.Uint16(b)
}
func Bytes3toInt32(b []uint8) uint32 {
	bytes := make([]byte, 1, 4)
	bytes = append(bytes, b...)
	return binary.BigEndian.Uint32(bytes)
}
func Bytes4toInt32(b []uint8) uint32 {
	return binary.BigEndian.Uint32(b)
}
func DecEmp(b uint8) (empresa uint8, catalogo uint8, size_emp uint8, pos_loc uint8, byte_size uint8) {

	// EMPRESA 4 - CATALOGO 2 - SIZE EMP 3 - POS LOC 3 - BYTE SIZE 3
	empresa = b / 54 // 4
	aux0 := b % 54
	catalogo = aux0 / 27 // 2
	aux1 := b % 27
	size_emp = aux1 / 9 // 3
	aux2 := b % 9
	pos_loc = aux2 / 3 // 3
	aux3 := b % 3
	byte_size = aux3 % 3 // 3
	return
}
func Float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}
func Distance(lat1, lng1, lat2, lng2 float32) uint32 {
	first := math.Pow(float64(lat2-lat1), 2)
	second := math.Pow(float64(lng2-lng1), 2)
	return uint32(math.Sqrt(first + second))
}
func DecProd(b uint8) (Tipo uint8, CantPrecio uint8, CantId uint8, Eval uint8, Filtro uint8, TipoId uint8) {

	// TIPO 2 - CANTPRECIO 4 - CANTID 3 - EVAL 2 - FILTRO 2 - TIPOID 2
	Tipo = b / 96
	aux1 := b % 96
	CantPrecio = aux1 / 24
	aux2 := aux1 % 24
	CantId = aux2 / 8
	aux3 := aux2 % 8
	Eval = aux3 / 4
	aux4 := aux3 % 4
	Filtro = aux4 / 2
	aux5 := aux4 % 2
	TipoId = aux5 % 2
	return
}
func DecodeSpecialBytes(byte []byte, limit int) (int, int) {
	m := int(byte[0])
	if m < limit {
		return m, 1
	} else {
		return (m-limit)*256 + int(byte[1]) + 200, 2
	}
}
func CompareFiltro(v []byte, p [][]uint32) float64 {

	var res float64 = 0
	var length int = len(v)
	var j int = 0
	for {

		Num, CantBytes, Tipo := DecFiltro(v[j])
		if Tipo == 0 {
			if len(p[Num]) == 1 {
				if CantBytes == 0 && p[Num][0] < 256 {
					if uint8(p[Num][0]) == v[j+1 : j+int(CantBytes)+2][0] {
						res = res + 100
					}
				}
				if CantBytes == 1 {
					if uint16(p[Num][0]) == Bytes2toInt16(v[j+1:j+int(CantBytes)+2]) {
						res = res + 100
					}
				}
			}
			j = j + int(CantBytes) + 2
		}

		if length <= j {
			break
		}
	}
	return res
}
func CompareEval(v []byte, p []uint8, suma uint32) uint32 {

	var sum uint32 = 0
	for i, x := range p {
		sum = sum + uint32(v[i]*x)
	}
	return sum * 100 / suma
}
func GetNotaMenor(Prods []ResProd, Nota float64) (float64, int, uint32) {

	var Posicion int = -1
	var PosEmp uint32 = 0
	for i, v := range Prods {
		if v.Nota < Nota {
			Nota = v.Nota
			Posicion = i
			PosEmp = v.IdEmp
		}
	}
	return Nota, Posicion, PosEmp
}
func DecProMem(b uint8) (CantString uint8, ProF uint8, ProE uint8) {

	CantString = b / 4
	aux1 := b % 4
	ProF = aux1 / 2
	aux2 := b % 2
	ProE = aux2 % 2
	return
}
func DecFiltro(b uint8) (Num uint8, CantBytes uint8, Tipo uint8) {

	Num = b / 4
	aux1 := b % 4
	CantBytes = aux1 / 2
	aux2 := aux1 % 2
	Tipo = aux2 % 2
	return
}
func Silence(s string, i int) {}

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
