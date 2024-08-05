package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

func buildClientConfig(clibanaConfig ClibanaConfig) opensearch.Config {
	var opensearchConfig opensearch.Config

	switch strings.ToLower(clibanaConfig.AuthType) {
	case AuthTypeAWS:
		opensearchConfig = buildAWSAuthClientConfig()
	case AuthTypeBasic:
		opensearchConfig = buildBasicAuthClientConfig(clibanaConfig)
	default:
		FatalError(fmt.Errorf("unsupported authentication type: %s", clibanaConfig.AuthType))
	}

	opensearchConfig.Addresses = []string{clibanaConfig.Host}
	opensearchConfig.Transport = &http.Transport{
		ResponseHeaderTimeout: ResponseTimeout * time.Second,
	}

	return opensearchConfig
}

func createClient(clibanaConfig ClibanaConfig) (*opensearch.Client, error) {
	if strings.HasPrefix(clibanaConfig.Host, "aws://") {
		domainName := strings.TrimPrefix(clibanaConfig.Host, "aws://")
		clibanaConfig.Host = resolveAWSDomainEndpoint(domainName)
	}

	return opensearch.NewClient(buildClientConfig(clibanaConfig))
}

func buildBasicAuthClientConfig(config ClibanaConfig) opensearch.Config {
	if config.Username == "" || config.Password == "" {
		FatalError(fmt.Errorf("uusername and password must be provided for basic authentication"))
	}

	return opensearch.Config{
		Username: config.Username,
		Password: config.Password,
	}
}

func doRequest[T any](client *opensearch.Client, request opensearchapi.Request) T {
	DebugLogger.Printf("Request: %+v\n", request)

	response, err := request.Do(context.TODO(), client)
	if err != nil {
		FatalError(fmt.Errorf("request failed: %w", err))
	}

	defer response.Body.Close()

	DebugLogger.Printf("Response: %+v\n", response)

	if response.IsError() {
		FatalError(fmt.Errorf("request error: %s", response.String()))
	}

	var responseObj T

	if err := json.NewDecoder(response.Body).Decode(&responseObj); err != nil {
		FatalError(fmt.Errorf("error decoding response: %w", err))
	}

	return responseObj
}
