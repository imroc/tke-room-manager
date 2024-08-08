package app

import (
	"github.com/imroc/tke-room-manager/pkg/prom"
)

func runPromApi() {
	if err := prom.StartServer(":9090"); err != nil {
		panic(err)
	}
}
