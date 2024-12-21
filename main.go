package main

import (
	"BSM/initializers"
	"BSM/routes"
)

func init() {
	initializers.LoadEnvVar()
	initializers.ConnectDB()
}
func main() {

	r := routes.SetupRoutes()
	r.Run()
}
