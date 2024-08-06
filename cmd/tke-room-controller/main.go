package main

import "github.com/imroc/tke-room-manager/cmd/tke-room-controller/app"

func main() {
	if err := app.RootCommand.Execute(); err != nil {
		panic(err)
	}
}
