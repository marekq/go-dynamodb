package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// scan dynamodb
func scan_ddb(svc *dynamodb.DynamoDB, table_name string, limit int64) ([]map[string]*dynamodb.AttributeValue, error, int) {

	// set count to 0
	count := 0

	// set parameters for initial scan
	input := &dynamodb.ScanInput{
		TableName:              aws.String(table_name),
		ReturnConsumedCapacity: aws.String("TOTAL"),
		Limit:                  aws.Int64(limit),
	}

	// scan table
	res, err := svc.Scan(input)

	// print error
	if err != nil {
		return nil, err, count
	}

	// get last evaluated key and data
	lastEvaluatedKey := res.LastEvaluatedKey
	data := res.Items

	// increase counter
	count += int(*res.Count)

	// if last evaluated key found
	for lastEvaluatedKey != nil {

		input := &dynamodb.ScanInput{
			TableName:              aws.String(table_name),
			ReturnConsumedCapacity: aws.String("TOTAL"),
			Limit:                  aws.Int64(limit),
			ExclusiveStartKey:      lastEvaluatedKey,
		}

		// scan table
		res, err := svc.Scan(input)

		// print errors
		if err != nil {
			fmt.Println(err)
			return nil, err, count
		}

		// append to data
		data = append(data, res.Items...)

		// update counter
		count += int(*res.Count)
		fmt.Println(strconv.Itoa(count))

		// set last evaluated key
		lastEvaluatedKey = res.LastEvaluatedKey
	}

	// create file on /tmp
	file, err := os.Create("/tmp/result.csv")
	defer file.Close()

	// write to csv
	writer := csv.NewWriter(file)
	defer writer.Flush()

	return data, nil, count
}

// lambda handler
func handler() {

	// set table name and retrieve limit
	table := os.Getenv("ddb_table")
	limit := int64(1000)

	// setup session with DynamoDB
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess)

	// scan table
	scan_ddb(svc, table, limit)
}

// start handler
func main() {
	lambda.Start(handler)
}
