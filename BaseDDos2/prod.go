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
)

type DDoS struct {
	TiempoIp  int64    `json:"TiempoIp"`
	Ips       Ip       `json:"Ips"`
	Start     bool     `json:"Start"`
	Count     uint32   `json:"Count"`
	BlackList [][]byte `json:"BlackList"`
}
type Ip struct {
	Poderacion uint8        `json:"Poderacion"`
	Tiempo     uint16       `json:"Tiempo"`
	Bytes      map[uint8]Ip `json:"Bytes"`
}

type Config struct {
	Tiempo time.Duration `json:"Tiempo"`
}
type MyHandler struct {
	Conf Config `json:"Conf"`
	DDoS DDoS   `json:"DDoS"`
}

func main() {

	var port string
	if runtime.GOOS == "windows" {
		port = ":81"
	} else {
		port = ":8080"
	}

	pass := &MyHandler{
		DDoS: DDoS{Start: true, Ips: Ip{Bytes: make(map[uint8]Ip, 0)}},
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

			//ctx.RemoteAddr().String()
			if !h.DDoS.Start || CreateIp(&h.DDoS, &h.DDoS.Ips, GetIp(string(ctx.QueryArgs().Peek("ip"))), 0) {
				ctx.SetBody([]byte{91, 48, 46, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 49, 93})
			} else {
				fmt.Fprintf(ctx, "IP BLOCKEADA - ENVIAR ERROR:")
			}
			fmt.Println(h.DDoS.Ips)
			fmt.Println(h.DDoS.Count)
			/*
				if b1, f1 := h.DDoS.Ips.Bytes[127]; f1 {
					if b2, f2 := b1.Bytes[0]; f2 {
						if b3, f3 := b2.Bytes[0]; f3 {
							if b4, f4 := b3.Bytes[1]; f4 {
								fmt.Println(b4)
							}
						}
					}
				}
			*/

		case "/print":

			fmt.Println(h.DDoS.Ips)
			fmt.Fprintf(ctx, "Print")

		case "/save":

			now := time.Now()
			cantip := h.SaveIps(0, false)
			fmt.Printf("Ips guardadas %v en %v\n", cantip, time.Since(now))
			fmt.Fprintf(ctx, "Save")

		case "/start":

			h.StartDDos()
			fmt.Fprintf(ctx, "Start")

		case "/stop":

			h.StopDDos()
			fmt.Fprintf(ctx, "Stop")

		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
}

// DAEMON //
func (h *MyHandler) StartDaemon() {
	h.Conf.Tiempo = 10 * time.Second
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

// DDoS //
func (h *MyHandler) StartDDos() {
	h.DDoS.Start = true
	h.DDoS.TiempoIp = time.Now().UnixMilli()
}
func (h *MyHandler) StopDDos() {
	h.DDoS.Start = false
	h.DDoS.TiempoIp = 0
	h.DDoS.Ips = Ip{Bytes: make(map[uint8]Ip, 0)}
}
func (h *MyHandler) SaveIps(tipo int, verb bool) int {

	if tipo == 0 {
		n := 5000000
		for i := 0; i < n; i++ {
			ip := newIntToBytes4byte(i)
			CreateIp(&h.DDoS, &h.DDoS.Ips, ip, 0)
			if verb {
				fmt.Println("Ip guardada: ", ip)
			}
		}
		return n
	}
	if tipo == 1 {
		z := 0
		ip := make([]byte, 4)
		ip[2] = 0
		ip[3] = 0
		for i := uint8(0); i <= 255; i++ {
			ip[0] = i
			for j := uint8(0); j <= 255; j++ {
				ip[1] = j
				for k := uint8(0); k <= 255; k++ {
					ip[2] = k
					for m := uint8(0); m <= 255; m++ {
						ip[3] = m

					}
				}
				CreateIp(&h.DDoS, &h.DDoS.Ips, ip, 0)
				z++
				if verb {
					fmt.Println("Ip guardada: ", ip)
				}
			}
		}
		return z
	}
	return 0
}
func ParamInt(data string) uint8 {
	var x uint8
	for _, c := range data {
		x = x*10 + uint8(c-'0')
	}
	return x
}
func verIp(Pip *Ip, ip []byte, i int, start int64) (bool, uint8) {
	if Ip, found := Pip.Bytes[ip[i]]; found {
		if i == 3 {
			return true, Ip.Poderacion
		} else {
			return verIp(&Ip, ip, i+1, start)
		}
	} else {
		return false, 0
	}
}
func GetIp(ip string) []byte {
	b := make([]byte, 4)
	var min, cant, j int = 0, 0, 0
	for _, i := range ip {
		if i == 46 || i == 58 {
			b[j] = ParamInt(ip[min:cant])
			if i == 58 {
				return b
			} else {
				min = cant + 1
				j++
			}
		}
		cant++
	}
	return b
}
func Ponderacion(p uint8, t uint16) uint8 {
	return p * p
}
func CreateIp(ddos *DDoS, ips *Ip, ip []uint8, i int) bool {
	if LocalIp, found := ips.Bytes[ip[i]]; found {
		if i < 3 {
			return CreateIp(ddos, &LocalIp, ip, i+1)
		} else {
			lapsed := uint16(time.Now().UnixMilli() - ddos.TiempoIp)
			LocalIp.Poderacion = Ponderacion(LocalIp.Poderacion, lapsed-LocalIp.Tiempo)
			LocalIp.Tiempo = lapsed
			if LocalIp.Poderacion < 255 {
				ddos.BlackList = append(ddos.BlackList, ip)
				return true
			} else {
				return false
			}
		}
	} else {
		al := Ip{}
		if i < 3 {
			al.Bytes = make(map[uint8]Ip, 0)
			ips.Bytes[ip[i]] = al
			x := ips.Bytes[ip[i]]
			return CreateIp(ddos, &x, ip, i+1)
		} else {
			ddos.Count++

			al.Tiempo = uint16(time.Now().UnixMilli() - ddos.TiempoIp)
			ips.Bytes[ip[i]] = al
			return true
		}
	}
}
func newIntToBytes4byte(num int) []byte {

	b := make([]byte, 4)
	var r int = num % 16777216
	b[0] = uint8(num / 16777216)
	b[1] = uint8(r / 65536)
	r = r % 65536
	b[2], b[3] = uint8(r/256), uint8(r%256)

	return b
}
