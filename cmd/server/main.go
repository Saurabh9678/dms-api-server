package main

import "infiour.local/dms-api-server/internal/bootstrap"

func main() {
	bootstrap.NewApp().Run()
}
