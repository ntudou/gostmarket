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

func post(c *Config) {
	si := new(ServiceInfo)
	si.Event = EVENT_PUSH
	si.Code = c.Code
	si.Namespace = c.Namespace
	itf, err := net.InterfaceByName(c.Interface)
	if err != nil {
		log.Println(err.Error())
		return
	}
	list, err := itf.Addrs()
	for _, v := range list {
		if ipnet, ok := v.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				si.Interfaces = ipnet.IP.String()
				break
			}
		}
	}
	si.Port = c.Port
	si.Stat = "ok"
	si.Active = time.Now().Unix()

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
	addr, err := net.ResolveUDPAddr("udp", c.Target)
	if err != nil {
		log.Println(err.Error())
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer conn.Close()
	cs := make([]byte, 4)
	checksum := crc32.ChecksumIEEE(encode)
	binary.BigEndian.PutUint32(cs, checksum)
	data := append(cs, encode...)
	_, err = conn.Write(data)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func get() {

}

func main() {
	myApp := cli.NewApp()
	myApp.Name = "gost market name service"
	myApp.Usage = "app client"
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
			Name:  "namespace,n",
			Value: "",
			Usage: "service namespace",
		},
		cli.StringFlag{
			Name:  "interface,i",
			Value: "eth0",
			Usage: "network interface",
		},
		cli.StringFlag{
			Name:  "port,p",
			Value: "",
			Usage: "service port",
		},
		cli.StringFlag{
			Name:  "code",
			Value: "",
			Usage: "instance global code",
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
		config.Mtu = c.Int64("mtu")
		config.Namespace = c.String("namespace")
		config.Interface = c.String("interface")
		config.Port = c.String("port")
		config.Code = c.String("code")
		AES_KEY = []byte(config.Key)

		if c.String("c") != "" {
			err := parseJSONConfig(config, c.String("c"))
			if err != nil {
				log.Fatalln(err.Error())
				return err
			}
		}

		for {
			post(config)
			time.Sleep(time.Minute)
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
