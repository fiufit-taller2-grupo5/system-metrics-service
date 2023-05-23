package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

type MetricsController struct {
	mongoClient *MongoClient
}

func NewMetricsController() (*MetricsController, error) {
	mongoClient, err := NewMongoClient()
	if err != nil {
		return nil, err
	}

	return &MetricsController{mongoClient: mongoClient}, nil
}

func (controller *MetricsController) SetUp(router gin.IRouter) {
	router.GET("/api/metrics", controller.GetMetric)
}

func (controller *MetricsController) GetMetric(c *gin.Context) {
	metric, errMetric := getNameQueryParam(c)
	if errMetric != nil {
		fmt.Println(errMetric.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMetric.Error()})
		return
	}

	interval := getIntervalFromQueryParam(c)
	from, to, errTime := getFromAndToFromQueryParams(c)

	numberOfDataPoints := bucketCountFromIntervalAndTimeSlice(*from, *to, interval)

	if errTime != nil {
		fmt.Println(errMetric.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errTime.Error()})
		return
	}

	metricDocuments, mongoErr := controller.getDocsByMetricAndTimeBounds(*metric, *from, *to)
	if mongoErr != nil {
		fmt.Println(errMetric.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": mongoErr.Error()})
		return
	}

	dataPoints := aggregateMetricsByBucket(metricDocuments, numberOfDataPoints, *from, *to)

	c.JSON(http.StatusOK, dataPoints)
}

func bucketCountFromIntervalAndTimeSlice(from, to time.Time, interval string) int {
	var divisor int64

	switch interval {
	case "minutes":
		divisor = int64(time.Minute)
	case "hours":
		divisor = int64(time.Hour)
	case "days":
		divisor = int64(24 * time.Hour)
	case "weeks":
		divisor = int64(7 * 24 * time.Hour)
	case "months":
		divisor = int64(30 * 24 * time.Hour) // using a 30-day approximation for simplicity
	case "years":
		divisor = int64(365 * 24 * time.Hour) // using a 365-day approximation for simplicity
	default:
		fmt.Println("Using default interval")
		return 10 // default interval
	}

	duration := to.Sub(from)
	if duration < 0 {
		duration = -duration
	}

	bucketCount := int(duration.Nanoseconds()/divisor) + 1
	return bucketCount
}

func getFromAndToFromQueryParams(c *gin.Context) (*time.Time, *time.Time, error) {
	from, fromExists := c.Get("from")
	to, toExists := c.Get("to")
	if !fromExists || !toExists || from == "" || to == "" {
		return nil, nil, errors.New("'from' and/or 'to' not specified or bad formatted")
	}

	fromTime, fromTimeErr := time.Parse(time.RFC3339, from.(string))
	toTime, toTimeErr := time.Parse(time.RFC3339, to.(string))
	if fromTimeErr != nil || toTimeErr != nil {
		return nil, nil, errors.New("'from' and/or 'to' have bad timestamp format")
	}

	return &fromTime, &toTime, nil
}

func getIntervalFromQueryParam(c *gin.Context) string {
	intervalQueryParam := c.Query("interval")

	if !validInterval(intervalQueryParam) {
		return "days"
	}

	return strings.ToLower(intervalQueryParam)
}

func validInterval(interval string) bool {
	interval = strings.ToLower(interval)
	return interval == "minutes" || interval == "hours" || interval == "days" || interval == "weeks" || interval == "months" || interval == "years"
}

func getNameQueryParam(c *gin.Context) (*string, error) {
	metric := c.Query("metric")
	if metric == "" {
		return nil, errors.New("'metric' query param not specified")
	}

	return &metric, nil
}
