package ec

import (
	"fmt"

	"github.com/crossplane/terrajet/pkg/config"
)

const (
	deployment    = "ec_deployment"
	elasticsearch = "elasticsearch"

	esUsername    = "elasticsearch_username"
	esPassword    = "elasticsearch_password"
	httpEndpoint  = "http_endpoint"
	httpsEndpoint = "https_endpoint"
	endpointFmt   = "%s_%d"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator(deployment, func(r *config.Resource) {
		r.ExternalName = config.IdentifierFromProvider

		r.Sensitive = config.Sensitive{
			AdditionalConnectionDetailsFn: func(attr map[string]interface{}) (map[string][]byte, error) {
				conn := map[string][]byte{}
				if a, ok := attr[esUsername].(string); ok {
					conn[esUsername] = []byte(a)
				}
				if a, ok := attr[esPassword].(string); ok {
					conn[esPassword] = []byte(a)
				}

				if a, ok := attr[elasticsearch].([]interface{}); ok {
					for i, nestedAttr := range a {
						if nestedMap, ok := nestedAttr.(map[string]interface{}); ok {
							if httpEndpoint, ok := nestedMap[httpEndpoint].(string); ok {
								conn[fmt.Sprintf(endpointFmt, httpEndpoint, i)] = []byte(httpEndpoint)
							}
							if httpsEndpoint, ok := nestedMap[httpsEndpoint].(string); ok {
								conn[fmt.Sprintf(endpointFmt, httpsEndpoint, i)] = []byte(httpsEndpoint)
							}
						}
					}
				}
				return conn, nil
			},
		}

		// Note(@tnthornton) deployments take more than 1 minute to provision.
		r.UseAsync = true
	})
}
