package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errMetric.Error()})
		return
	}

	numberOfDataPoints := getBucketsFromQueryParams(c)

	from, to, errTime := getFromAndToFromQueryParams(c)
	if errTime != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": errTime.Error()})
		return
	}

	metricDocuments, mongoErr := controller.getDocsByMetricAndTimeBounds(*metric, *from, *to)
	if mongoErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": mongoErr.Error()})
		return
	}

	dataPoints := aggregateMetricsByBucket(metricDocuments, numberOfDataPoints, *from, *to)

	c.JSON(http.StatusOK, dataPoints)
}

func getFromAndToFromQueryParams(c *gin.Context) (*time.Time, *time.Time, error) {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		return nil, nil, errors.New("'from' and/or 'to' not specified")
	}

	fromTime, fromTimeErr := time.Parse(time.RFC3339, from)
	toTime, toTimeErr := time.Parse(time.RFC3339, to)
	if fromTimeErr != nil || toTimeErr != nil {
		return nil, nil, errors.New("'from' and/or 'to' have bad timestamp format")
	}

	return &fromTime, &toTime, nil
}

func getBucketsFromQueryParams(c *gin.Context) int {
	bucketsQueryParam := c.Query("buckets")
	bucketsAmount, err := strconv.Atoi(bucketsQueryParam)
	if err != nil || bucketsAmount < 3 {
		return 10
	}

	return bucketsAmount
}

func getNameQueryParam(c *gin.Context) (*string, error) {
	metric := c.Query("metric")
	if metric == "" {
		return nil, errors.New("'metric' query param not specified")
	}

	return &metric, nil
}
