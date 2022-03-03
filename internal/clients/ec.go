/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clients

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/terrajet/pkg/terraform"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-jet-ec/apis/v1alpha1"
)

const (
	keyAPIKey   = "apikey"
	keyUsername = "username"
	keyPassword = "password"
	keyHost     = "endpoint"

	// EC credentials environment variable names
	envAPIKey   = "EC_API_KEY"
	envUsername = "EC_USERNAME"
	envPassword = "EC_PASSWORD"
	envHost     = "endpoint"
)

const (
	fmtEnvVar = "%s=%s"

	// error messages
	errNoProviderConfig     = "no providerConfigRef provided"
	errGetProviderConfig    = "cannot get referenced ProviderConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal ec credentials as JSON"

	errEitherAPIKeyOrUP = "either apiKey OR username and password may be supplied"
	errUnameAndPword    = "username and password are required"
)

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		configRef := mg.GetProviderConfigReference()
		if configRef == nil {
			return ps, errors.New(errNoProviderConfig)
		}
		pc := &v1alpha1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
			return ps, errors.Wrap(err, errGetProviderConfig)
		}

		t := resource.NewProviderConfigUsageTracker(client, &v1alpha1.ProviderConfigUsage{})
		if err := t.Track(ctx, mg); err != nil {
			return ps, errors.Wrap(err, errTrackUsage)
		}

		data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, client, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}
		ecCreds := map[string]string{}
		if err := json.Unmarshal(data, &ecCreds); err != nil {
			return ps, errors.Wrap(err, errUnmarshalCredentials)
		}

		// set provider configuration
		ps.Configuration = map[string]interface{}{
			envHost: ecCreds[keyHost],
		}

		envVars, err := getAuthEnvVars(ecCreds)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}

		ps.Env = envVars

		return ps, nil
	}
}

func getAuthEnvVars(ecCreds map[string]string) ([]string, error) {
	// set environment variables for sensitive provider configuration
	// NOTE(@tnthornton) either apiKey OR username and password are valid
	// not both.
	apikey := ecCreds[keyAPIKey]
	uname := ecCreds[keyUsername]
	pword := ecCreds[keyPassword]

	if apikey != "" && (uname != "" || pword != "") {
		return nil, errors.New(errEitherAPIKeyOrUP)
	}

	if apikey != "" {
		return []string{
			fmt.Sprintf(fmtEnvVar, envAPIKey, apikey),
		}, nil
	}

	if uname == "" || pword == "" {
		return nil, errors.New(errUnameAndPword)
	}

	return []string{
		fmt.Sprintf(fmtEnvVar, envUsername, uname),
		fmt.Sprintf(fmtEnvVar, envPassword, pword),
	}, nil
}
