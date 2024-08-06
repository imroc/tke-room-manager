package main

import "github.com/imroc/tke-room-manager/cmd/app"

func main() {
	if err := app.RootCommand.Execute(); err != nil {
		panic(err)
	}
}
