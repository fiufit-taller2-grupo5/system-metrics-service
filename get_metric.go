package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type aggregatedMetricDocument struct {
	MetricName string              `json:"metric" bson:"metric"`
	Count      int                 `json:"count" bson:"count"`
	Timestamp  primitive.Timestamp `json:"timestamp" bson:"timestamp"`
}

func (controller *MetricsController) getDocsByMetricAndTimeBounds(metric string, from, to time.Time) ([]aggregatedMetricDocument, error) {
	mongoOptions := options.Find().SetProjection(bson.M{"_id": 0}) // Exclude _id from query (0 for excluding, 1 for including)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := controller.mongoClient.GetClient().Database("fiufit").Collection("system-metrics")

	filter := bson.M{"metric": metric}
	filter["$and"] = []bson.M{
		{"timestamp": bson.M{"$gte": primitive.Timestamp{T: uint32(from.Unix())}}},
		{"timestamp": bson.M{"$lte": primitive.Timestamp{T: uint32(to.Unix())}}},
	}

	cursor, err := collection.Find(ctx, filter, mongoOptions)
	if err != nil {
		return nil, err
	}

	var metricDocuments []aggregatedMetricDocument

	fmt.Printf("Batch size: %d\n", cursor.RemainingBatchLength())

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

func timestampToTime(ts primitive.Timestamp) time.Time {
	sec := int64(ts.T)
	nanoSeconds := int64(ts.I)
	return time.Unix(sec, nanoSeconds)
}

func aggregateMetricsByBucket(metrics []aggregatedMetricDocument, buckets int, from, to time.Time) []DataPoint {
	fmt.Printf("Metrics: %+v", metrics)

	duration := to.Sub(from)
	bucketDuration := (duration / time.Duration(buckets)).Round(time.Second)
	dataPoints := make([]DataPoint, buckets)
	for i := 0; i < buckets; i++ {
		startTime := from.Add(bucketDuration * time.Duration(i))
		endTime := startTime.Add(bucketDuration)

		count := 0
		for _, aggregatedMetric := range metrics {
			metricTimestamp := timestampToTime(aggregatedMetric.Timestamp)
			if metricTimestamp.After(startTime) && metricTimestamp.Before(endTime) {
				count += aggregatedMetric.Count
			}
		}

		dataPoints[i] = DataPoint{Start: startTime, End: endTime, Count: count, Position: i}
	}

	return dataPoints
}
