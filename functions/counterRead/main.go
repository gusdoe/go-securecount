package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const region = "us-east-1"
const counterTableName = "counter-dev"

// Counter structure
type Counter struct {
	Count int `json:"count"`
}

// Handler handles read counter requests
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	svc := dynamodb.New(sess)
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(counterTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"userId": {
				S: aws.String(request.RequestContext.Identity.CognitoIdentityID),
			},
		},
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	counter := Counter{
		Count: 0,
	}

	if result.Item == nil {
		ci, err := dynamodbattribute.MarshalMap(counter)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, err
		}
		ci["userId"] = &dynamodb.AttributeValue{
			S: aws.String(request.RequestContext.Identity.CognitoIdentityID),
		}
		input := &dynamodb.PutItemInput{
			TableName: aws.String(counterTableName),
			Item:      ci,
		}

		output, err := svc.PutItem(input)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, err
		}
		result.Item = output.Attributes
	} else {
		err = dynamodbattribute.UnmarshalMap(result.Item, counter)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
			}, err
		}
	}

	counterJSON, err := json.Marshal(counter)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(counterJSON),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
