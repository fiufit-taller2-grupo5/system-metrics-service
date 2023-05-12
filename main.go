package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Hello, world!")

	router := gin.Default()

	metricsController, err := NewMetricsController()
	if err != nil {
		fmt.Println("Could not start metrics controller: " + err.Error())
	}

	metricsController.SetUp(router)

	err = router.Run(":8008")
	if err != nil {
		fmt.Println("Server failed running: " + err.Error())
	}
}
