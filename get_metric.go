package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type aggregatedMetricDocument struct {
	MetricName string    `json:"metric"`
	Count      int       `json:"count"`
	Timestamp  time.Time `json:"timestamp"`
}

func (controller *MetricsController) getDocsByMetricAndTimeBounds(metric string, from, to time.Time) ([]aggregatedMetricDocument, error) {
	mongoOptions := options.Find().SetProjection(bson.M{"_id": 0})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := controller.mongoClient.GetClient().Database("fiufit").Collection("system-metrics")

	filter := bson.M{"metric": metric}
	filter["$and"] = []bson.M{
		{"timestamp": bson.M{"$gte": from}},
		{"timestamp": bson.M{"lte": to}},
	}

	cursor, err := collection.Find(ctx, filter, mongoOptions)
	if err != nil {
		return nil, err
	}

	var metricDocuments []aggregatedMetricDocument
	for cursor.Next(ctx) {
		var document aggregatedMetricDocument
		decodeErr := cursor.Decode(&document)
		if decodeErr != nil {
			return nil, decodeErr
		}

		metricDocuments = append(metricDocuments, document)
	}

	return metricDocuments, nil
}

type DataPoint struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Count    int       `json:"count"`
	Position int       `json:"position"`
}

func aggregateMetricsByBucket(metrics []aggregatedMetricDocument, buckets int, from, to time.Time) []DataPoint {
	duration := to.Sub(from)
	bucketDuration := (duration / time.Duration(buckets)).Round(time.Second)
	dataPoints := make([]DataPoint, buckets)
	for i := 0; i < buckets; i++ {
		startTime := from.Add(bucketDuration * time.Duration(i))
		endTime := startTime.Add(bucketDuration)

		count := 0
		for _, aggregatedMetric := range metrics {
			if aggregatedMetric.Timestamp.After(startTime) && aggregatedMetric.Timestamp.Before(endTime) {
				count += aggregatedMetric.Count
			}
		}

		dataPoints[i] = DataPoint{Start: startTime, End: endTime, Count: count, Position: i}
	}

	return dataPoints
}
