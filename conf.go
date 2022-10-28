package main

import (
	"encoding/json"
	"os"
)

type config struct {
	Debug       bool   `json:"debug"`
	Port        uint32 `json:"port"`
	RoutersFile string `json:"routers_file"`
}

func getConfig() *config {
	c := &config{
		Debug:       false,
		Port:        8899,
		RoutersFile: "routers.json",
	}
	b, err := os.ReadFile("conf.json")
	if err != nil {
		return c
	}
	err = json.Unmarshal(b, c)
	if err != nil {
		return c
	}
	return c
}
