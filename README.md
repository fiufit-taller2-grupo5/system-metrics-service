# System Metrics Service

## Description

This service is a HTTP server responsible for serving the aggregated metrics to the frontend, to use them in plots or in any analysis needed.

It fetches the data directly from Mongo Atlas, so the credentials are needed to work, even if you are running it locally.

It uses Gin as HTTP framework

## Running the project

First, make sure the Mongo Atlas credentials are set correctly. To run the server, run `go get -u` to update the dependencies, and `go run main.go` to run the server.

## Endpoints

The only endpoint is a GET to `api/metrics`. It needs some query params to get the query configuration:

1. `metric`: The metric name to get, its the same name than stored in mongo. E.g: `user-created`
2. `interval`: The grouping interval of the data points. It can be: `minutes`, `hours`, `days`, `weeks`, `months`, `years`. The default value is `days`, if not specified.
3. `from` and `to`: The date range for the data points. Is a string with format RFC3339.

## How does it work

Once a request is received, it parses all query params, and calculates the amount of Data Points and their date ranges. Then, it fetches from mongo all the ocurrences of the desired metric in the specified time range, and inserts each one into their bucket (data point). It then returns an array of Data Points jsons of the form:

```json
{
  "start": "2020-01-01T00:00:00Z",
  "end": "2020-01-02T00:00:00Z",
  "count": 10,
  "position": 1 // The position in the array
}
```

Good Luck!
