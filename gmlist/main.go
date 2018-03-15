package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli"
)

const (
	// VERSION is injected by buildflags
	VERSION = "SELFBUILD"
)

func get(c *Config) {
	resp, err := http.Get("http://" + c.Target)
	if err != nil {
		log.Println(err.Error())
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
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
			Name:  "c",
			Value: "", // when the value is not empty, the config path must exists
			Usage: "config from json file, which will override the command from shell",
		},
	}

	myApp.Action = func(c *cli.Context) error {
		config := new(Config)
		config.Target = c.String("target")
		config.Key = c.String("key")
		AES_KEY = []byte(config.Key)

		if c.String("c") != "" {
			err := parseJSONConfig(config, c.String("c"))
			if err != nil {
				log.Fatalln(err.Error())
				return err
			}
		}

		for {
			get(config)
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
