package oci

import (
	"context"
	"fmt"

	"encoding/base64"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/secrets"
)

type OciClient interface {
	GetSecret(ctx context.Context, secretId string) ([]byte, error)
}

type ociClient struct {
	config  common.ConfigurationProvider
	secrets secrets.SecretsClient
}

func NewClient() (OciClient, error) {

	return NewClientWithConfiguration(nil, nil)
}

func NewClientWithConfiguration(path *string, profile *string) (OciClient, error) {

	var err error
	var config common.ConfigurationProvider

	if profile == nil {
		config, err = auth.InstancePrincipalConfigurationProvider()

	} else {
		config = common.CustomProfileConfigProvider(*path, *profile)
	}

	if err != nil {
		return nil, fmt.Errorf("invalid OCI client configuration: %v", err)
	}

	fmt.Printf("config: %v", config)

	client, err := secrets.NewSecretsClientWithConfigurationProvider(config)

	if err != nil {
		return nil, fmt.Errorf("cannot create the OCI client: %v", err)
	}

	return &ociClient{
		config:  config,
		secrets: client,
	}, nil
}

func (c *ociClient) GetSecret(ctx context.Context, secretId string) ([]byte, error) {

	bundle, err := c.secrets.GetSecretBundle(ctx, secrets.GetSecretBundleRequest{
		SecretId: &secretId,
	})

	if err != nil {
		return nil, fmt.Errorf("cannot load secret %s: %v", secretId, err)
	}

	content := bundle.SecretBundle.SecretBundleContent.(secrets.Base64SecretBundleContentDetails)

	secret, err := base64.StdEncoding.DecodeString(*content.Content)

	if err != nil {
		return nil, fmt.Errorf("cannot read secret %s: %v", secretId, err)
	}

	return secret, nil
}
