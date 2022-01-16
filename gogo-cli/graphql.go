package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Khan/genqlient/graphql"
)

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("x-hasura-admin-secret", t.key)
	return t.wrapped.RoundTrip(req)
}

func createGQLClient() graphql.Client {
	key := os.Getenv("HASURA_ADMIN_SECRET")
	if key == "" {
		err := fmt.Errorf("must set HASURA_ADMIN_SECRET")
		handleErr(err)
	}

	endpoint := os.Getenv("HASURA_ENDPOINT")

	if endpoint == "" {
		err := fmt.Errorf("must set HASURA_ENDPOINT")
		handleErr(err)
	}

	httpClient := http.Client{
		Transport: &authedTransport{
			key:     key,
			wrapped: http.DefaultTransport,
		},
	}
	return graphql.NewClient(endpoint, &httpClient)
}

// Main gql client
var graphqlClient = createGQLClient()
