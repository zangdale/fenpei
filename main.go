package main

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

var conf = getConfig()
var routers *routersConf
var log = zlog.Logger

func init() {
	routers = getRouters(conf.RoutersFile)
	if conf.Debug {
		log = log.Level(zerolog.ErrorLevel)
	}
	log.Info().Bool("debug", conf.Debug).Msg("log init")

}

func main() {
	hand := NewHandler(confToRouters(routers.Routers))
	go func() {
		for {
			if e := recover(); e != nil {
				buf := make([]byte, 1024*2)
				runtime.Stack(buf, true)
				log.Error().Interface("err", e).Bytes("body", buf).Msg("error panic")
			}

			log.Info().Uint32("port", routers.Port).Msg("start router http server")
			err := http.ListenAndServe(fmt.Sprintf(":%d", routers.Port), hand)
			if err != nil {
				log.Error().Err(err).Msg("router http stop")
			}
		}
	}()

	log.Info().Uint32("port", conf.Port).Msg("start control http server")
	err := http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), controler(hand))
	if err != nil {
		log.Error().Err(err).Msg("control http stop")
	}

}
