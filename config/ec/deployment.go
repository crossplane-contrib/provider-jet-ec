package ec

import (
	"fmt"

	"github.com/crossplane/terrajet/pkg/config"
)

// Configure configures individual resources by adding custom
// ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("ec_deployment", func(r *config.Resource) {
		r.ExternalName = config.IdentifierFromProvider

		r.Sensitive = config.Sensitive{
			AdditionalConnectionDetailsFn: func(attr map[string]interface{}) (map[string][]byte, error) {
				conn := map[string][]byte{}
				if a, ok := attr["elasticsearch_username"].(string); ok {
					conn["elasticsearch_username"] = []byte(a)
				}
				if a, ok := attr["elasticsearch_password"].(string); ok {
					conn["elasticsearch_password"] = []byte(a)
				}

				if a, ok := attr["elasticsearch"].([]interface{}); ok {
					for i, nestedAttr := range a {
						if nestedMap, ok := nestedAttr.(map[string]interface{}); ok {
							if httpEndpoint, ok := nestedMap["http_endpoint"].(string); ok {
								conn[fmt.Sprintf("http_endpoint_%d", i)] = []byte(httpEndpoint)
							}
							if httpsEndpoint, ok := nestedMap["https_endpoint"].(string); ok {
								conn[fmt.Sprintf("https_endpoint_%d", i)] = []byte(httpsEndpoint)
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
