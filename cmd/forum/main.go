package main

import "fmt"

func main() {
	service.InitDependencies()
	fmt.Println("Forum service is running...")
	service.StartServer()
}