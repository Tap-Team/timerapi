package app

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func profilier(pprofUrl string) {
	go func() {
		log.Println("profilier err: ", http.ListenAndServe(pprofUrl, nil))
	}()
}
