package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"strings"
)

func decodeIfNecessary(param string) (string, error) {
	if !strings.Contains(param, "%") {
		return param, nil
	}

	decoded, err := url.QueryUnescape(param)
	if err != nil {
		return "", err
	}

	return decoded, nil
}

func DateDecode() gin.HandlerFunc {
	return func(c *gin.Context) {
		from, fromErr := decodeIfNecessary(c.Query("from"))
		to, toErr := decodeIfNecessary(c.Query("to"))
		if fromErr != nil || toErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}

		fmt.Println("Value of from: " + from)
		fmt.Println("Value of to: " + to)

		c.Set("from", from)
		c.Set("to", to)
		c.Next()
	}
}
