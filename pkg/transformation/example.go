package main

import (
	"context"
	"encoding/json"
	"io"

	tranformation "github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	SessionID string
	Body      io.Reader
}

type Data struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, req *Request) (string, error) {
	var d []Data
	json.NewDecoder(req.Body).Decode(&d)
}

func main() {
	tranformation.Start(HandleRequest)
}
