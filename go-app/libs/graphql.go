package libs

import (
	"net/http"
	"os"

	"github.com/shurcooL/graphql"
)

type customTransport struct {
	transport http.RoundTripper
	headers   map[string]string
}

func (c *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add custom headers to the request
	for key, value := range c.headers {
		req.Header.Add(key, value)
	}
	return c.transport.RoundTrip(req)
}

func SetupGraphqlClient() *graphql.Client {
	httpClient := &http.Client{
		Transport: &customTransport{
			transport: http.DefaultTransport,
			headers: map[string]string{
				"X-Hasura-Admin-Secret": os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET"),
			},
		},
	}

	apiEndpoint := os.Getenv("HASURA_GRAPHQL_API_ENDPOINT")
	client := graphql.NewClient(apiEndpoint, httpClient)

	return client
}
