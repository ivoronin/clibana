package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awsopensearch "github.com/aws/aws-sdk-go-v2/service/opensearch"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/signer/awsv2"
)

var awsConfig *aws.Config

func awsInit() {
	if awsConfig == nil {
		config, err := awsconfig.LoadDefaultConfig(context.TODO())
		if err != nil {
			FatalError(fmt.Errorf("failed to load AWS config: %w", err))
		}

		awsConfig = &config
	}
}

func resolveAWSDomainEndpoint(domainName string) string {
	awsInit()

	aosClient := awsopensearch.NewFromConfig(*awsConfig)

	domain, err := aosClient.DescribeDomain(context.TODO(), &awsopensearch.DescribeDomainInput{
		DomainName: &domainName,
	})
	if err != nil {
		FatalError(fmt.Errorf("failed to describe OpenSearch domain: %w", err))
	}

	var endpoint string

	switch {
	case domain.DomainStatus.EndpointV2 != nil:
		endpoint = *domain.DomainStatus.EndpointV2
	case domain.DomainStatus.Endpoint != nil:
		endpoint = *domain.DomainStatus.Endpoint
	case domain.DomainStatus.Endpoints != nil:
		switch {
		case domain.DomainStatus.Endpoints["vpcv2"] != "":
			endpoint = domain.DomainStatus.Endpoints["vpcv2"]
		case domain.DomainStatus.Endpoints["vpc"] != "":
			endpoint = domain.DomainStatus.Endpoints["vpc"]
		}
	}

	if endpoint == "" {
		FatalError(fmt.Errorf("no endpoints found for OpenSearch domain: %s", domainName))
	}

	return "https://" + endpoint
}

func buildAWSAuthClientConfig() opensearch.Config {
	awsInit()

	signer, err := awsv2.NewSigner(*awsConfig)
	if err != nil {
		FatalError(fmt.Errorf("failed to create AWS V4 signer: %w", err))
	}

	return opensearch.Config{
		Signer: signer,
	}
}
