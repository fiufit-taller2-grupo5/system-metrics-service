package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
)

func main() {
	router := gin.Default()

	metricsController, err := NewMetricsController()
	if err != nil {
		fmt.Println("Could not start metrics controller: " + err.Error())
	}

	metricsController.SetUp(router)

	port := portFromEnv()

	err = router.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Server failed running: " + err.Error())
	}
}

func portFromEnv() int {
	if os.Getenv("environment") == "production" {
		return 80
	} else {
		return 8006
	}
}
