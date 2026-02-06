package main

import  (
	"forum/internal/service"
	"fmt"
)

func main() {
	service.InitDependencies()
	fmt.Println("Forum service is running...")
	service.StartServer()
}