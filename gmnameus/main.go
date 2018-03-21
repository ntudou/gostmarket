package main

import (
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"log"
	"net"
	"os"
	"sync"

	"github.com/urfave/cli"
)

const (
	// VERSION is injected by buildflags
	VERSION    = "SELFBUILD"
	EVENT_END  = 0xFF
	EVENT_PUSH = 0x00
	EVENT_PULL = 0x01
	EVENT_GET  = 0x02
	EVENT_LIST = 0x03
)

var service_map sync.Map

func eachread(conn *net.UDPConn, addr *net.UDPAddr, data []byte, n int) {
	if n < 4 {
		log.Println("data len is", n)
		sendend(addr)
		return
	}
	checksum := data[0:4]
	body := data[4:n]
	cs := crc32.ChecksumIEEE(body)
	if cs != binary.BigEndian.Uint32(checksum) {
		log.Println("checksum err")
		sendend(addr)
		return
	}
	decode, err := GCMDecrypt(body)
	if err != nil {
		log.Println(err.Error())
		sendend(addr)
		return
	}
	si := new(ServiceInfo)
	err = json.Unmarshal(decode, si)
	if err != nil {
		log.Println(err.Error())
		sendend(addr)
		return
	}
	switch si.Event {
	case EVENT_PUSH:
		si.Host = addr.String()
		if _, ok := service_map.Load(si.Code); ok {
			service_map.Delete(si.Code)
		}
		service_map.Store(si.Code, si)
	case EVENT_PULL:
		namespace := si.Namespace
		sm := make(map[string]*ServiceInfo)
		service_map.Range(func(k, v interface{}) bool {
			s := v.(*ServiceInfo)
			if s.Namespace == namespace {
				sm[k.(string)] = s
			}
			return true
		})
		body, err := json.Marshal(sm)
		if err != nil {
			log.Println(err.Error())
			sendend(addr)
			return
		}
		sendevent(addr, body)
	case EVENT_GET:
		code := si.Code
		s, ok := service_map.Load(code)
		if !ok {
			log.Println(si.Code, "not found")
			sendend(addr)
			return
		}
		body, err := json.Marshal(s)
		if err != nil {
			log.Println(err.Error())
			sendend(addr)
			return
		}
		sendevent(addr, body)
	case EVENT_LIST:
		sm := make(map[string]*ServiceInfo)
		service_map.Range(func(k, v interface{}) bool {
			sm[k.(string)] = v.(*ServiceInfo)
			return true
		})
		body, err := json.Marshal(sm)
		if err != nil {
			log.Println(err.Error())
			sendend(addr)
			return
		}
		log.Println(addr)
		sendevent(addr, body)
	case EVENT_END:
		return
	}
}

func sendend(addr *net.UDPAddr) {
	si := new(ServiceInfo)
	si.Event = EVENT_END
	body, err := json.Marshal(si)
	if err != nil {
		log.Println(err.Error())
		return
	}
	sendevent(addr, body)
}

func sendevent(addr *net.UDPAddr, body []byte) {
	encode, err := GCMEncrypt(body)
	if err != nil {
		log.Println(err.Error())
		return
	}
	cs := make([]byte, 4)
	checksum := crc32.ChecksumIEEE(encode)
	binary.BigEndian.PutUint32(cs, checksum)
	data := append(cs, encode...)
	remoteaddr, err := net.ResolveUDPAddr("udp", addr.IP.String()+":10800")
	if err != nil {
		log.Fatalln(err.Error())
	}
	conn, err := net.DialUDP("udp", nil, remoteaddr)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer conn.Close()
	n, err := conn.Write(data)
	log.Println(n)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func main() {
	myApp := cli.NewApp()
	myApp.Name = "gost market name service"
	myApp.Usage = "app server"
	myApp.Version = VERSION
	myApp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "listen,l",
			Value: ":8802",
			Usage: "local listen address",
		},
		cli.StringFlag{
			Name:  "key,k",
			Value: "12345678901234567890123456789012",
			Usage: "key",
		},
		cli.Int64Flag{
			Name:  "mtu,m",
			Value: 1350,
			Usage: "set maximum transmission unit for UDP packets",
		},
		cli.StringFlag{
			Name:  "c",
			Value: "", // when the value is not empty, the config path must exists
			Usage: "config from json file, which will override the command from shell",
		},
	}
	myApp.Action = func(c *cli.Context) error {
		config := new(Config)
		config.Listen = c.String("listen")
		config.Key = c.String("key")
		config.Mtu = c.Int64("mtu")
		AES_KEY = []byte(config.Key)
		if c.String("c") != "" {
			err := parseJSONConfig(config, c.String("c"))
			if err != nil {
				log.Fatalln(err.Error())
				return err
			}
		}
		addr, err := net.ResolveUDPAddr("udp", config.Listen)
		if err != nil {
			log.Fatalln(err.Error())
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer conn.Close()
		for {
			data := make([]byte, config.Mtu)
			n, remoteaddr, err := conn.ReadFromUDP(data)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			go eachread(conn, remoteaddr, data, n)
		}
		return nil
	}
	myApp.Run(os.Args)
}

func init() {
	if VERSION == "SELFBUILD" {
		// add more log flags for debugging
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}
}
