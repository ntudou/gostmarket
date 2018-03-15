package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/urfave/cli"
)

const (
	// VERSION is injected by buildflags
	VERSION = "SELFBUILD"
)

var service_map sync.Map

func HttpRnd(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Method == "GET" {
		if len(r.Form["c"]) > 0 {
			code := r.Form["c"][0]
			s, ok := service_map.Load(code)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			body, err := json.Marshal(s)
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			encode, err := GCMEncrypt(body)
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(encode)
		} else if len(r.Form["n"]) > 0 {
			namespace := r.Form["n"][0]
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
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			encode, err := GCMEncrypt(body)
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(encode)
		} else {
			sm := make(map[string]*ServiceInfo)
			service_map.Range(func(k, v interface{}) bool {
				sm[k.(string)] = v.(*ServiceInfo)
				return true
			})
			body, err := json.Marshal(sm)
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			encode, err := GCMEncrypt(body)
			if err != nil {
				log.Println(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(encode)
		}
	} else if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		decode, err := GCMDecrypt(body)
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		si := new(ServiceInfo)
		err = json.Unmarshal(decode, si)
		if err != nil {
			log.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		si.Host = r.RemoteAddr
		if _, ok := service_map.Load(si.Code); ok {
			service_map.Delete(si.Code)
		}
		service_map.Store(si.Code, si)
	}
}

func main() {
	myApp := cli.NewApp()
	myApp.Name = "gost market name service"
	myApp.Usage = "http server"
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
		cli.StringFlag{
			Name:  "c",
			Value: "", // when the value is not empty, the config path must exists
			Usage: "config from json file, which will override the command from shell",
		},
	}
	myApp.Action = func(c *cli.Context) error {
		config := Config{}
		config.Listen = c.String("listen")
		config.Key = c.String("key")
		AES_KEY = []byte(config.Key)
		if c.String("c") != "" {
			err := parseJSONConfig(&config, c.String("c"))
			if err != nil {
				log.Fatalln(err.Error())
				return err
			}
		}

		http.HandleFunc("/", HttpRnd)
		err := http.ListenAndServe(config.Listen, nil)
		if err != nil {
			log.Fatalln(err.Error())
			return err
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
