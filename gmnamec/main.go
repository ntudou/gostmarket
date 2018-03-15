package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli"
)

const (
	// VERSION is injected by buildflags
	VERSION = "SELFBUILD"
)

func post(c *Config) {
	si := new(ServiceInfo)
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
	data := bytes.NewBuffer(encode)
	_, err = http.Post("http://"+c.Target, "application/json;charset=utf-8", data)
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
	myApp.Usage = "http server"
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
