package main

import (
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"log"
	"net"
	"os"
	"time"

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

func eachwrite(addr *net.UDPAddr, config *Config) {
	time.Sleep(time.Second)
	si := new(ServiceInfo)
	si.Event = EVENT_LIST
	body, err := json.Marshal(si)
	if err != nil {
		log.Println(err.Error())
		return
	}
	encode, err := GCMEncrypt(body)
	if err != nil {
		log.Println(err.Error())
		return
	}
	cs := make([]byte, 4)
	checksum := crc32.ChecksumIEEE(encode)
	binary.BigEndian.PutUint32(cs, checksum)
	data := append(cs, encode...)
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer conn.Close()
	log.Println(addr)
	_, err = conn.Write(data)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func eachread(data []byte, n int) {
	if n < 4 {
		log.Println("list len is ", n)
		return
	}
	var cs, body []byte
	cs = data[0:4]
	body = data[4:n]
	checksum := crc32.ChecksumIEEE(body)
	if checksum != binary.BigEndian.Uint32(cs) {
		log.Println("checksum err")
		return
	}
	decode, err := GCMDecrypt(body)
	if err != nil {
		log.Println(err.Error())
		return
	}
	service_map := make(map[string]*ServiceInfo)
	err = json.Unmarshal(decode, &service_map)
	if err != nil {
		log.Println(err.Error())
		return
	}
	for _, v := range service_map {
		log.Println(v)
	}
}

func main() {
	myApp := cli.NewApp()
	myApp.Name = "gost market name service"
	myApp.Usage = "app show"
	myApp.Version = VERSION
	myApp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "target,t",
			Value: "127.0.0.1:8802",
			Usage: "target server address",
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
		config.Target = c.String("target")
		config.Key = c.String("key")
		config.Mtu = c.Int64("mut")
		AES_KEY = []byte(config.Key)

		if c.String("c") != "" {
			err := parseJSONConfig(config, c.String("c"))
			if err != nil {
				log.Fatalln(err.Error())
				return err
			}
		}

		addr, err := net.ResolveUDPAddr("udp", config.Target)
		if err != nil {
			log.Fatalln(err.Error())
		}
		local, err := net.ResolveUDPAddr("udp", ":10800")
		if err != nil {
			log.Fatalln(err.Error())
		}
		conn, err := net.ListenUDP("udp", local)
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer conn.Close()
		for {
			go eachwrite(addr, config)
			data := make([]byte, config.Mtu)
			n, _, err := conn.ReadFromUDP(data)
			log.Println(n)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			go eachread(data, n)
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
