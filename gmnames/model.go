package main

type ServiceInfo struct {
	Code       string   `json:"code"`
	Namespace  string   `json:"namespace"`
	Interfaces string   `json:"interfaces"`
	Host       string   `json:"host"`
	Stat       string   `json:"stat"`
	Tag        []string `json:"tag"`
	Active     int64    `json:"active"`
}