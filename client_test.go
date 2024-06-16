package sms

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/hasura/go-graphql-client"
)

func cleanup(t *testing.T, client *Client) {

	_, err := client.CancelSms(map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
}

// hasuraTransport transport for Hasura GraphQL Client
type hasuraTransport struct {
	adminSecret string
	headers     map[string]string
	// keep a reference to the client's original transport
	rt http.RoundTripper
}

// RoundTrip set header data before executing http request
func (t *hasuraTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.adminSecret != "" {
		r.Header.Set("X-Hasura-Admin-Secret", t.adminSecret)
	}
	for k, v := range t.headers {
		r.Header.Set(k, v)
	}
	return t.rt.RoundTrip(r)
}

func newGqlClient() *graphql.Client {
	adminSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")
	httpClient := &http.Client{
		Transport: &hasuraTransport{
			rt:          http.DefaultTransport,
			adminSecret: adminSecret,
		},
		Timeout: 30 * time.Second,
	}
	return graphql.NewClient(os.Getenv("DATA_URL"), httpClient)
}

func TestSendSMSs(t *testing.T) {

	client := New(newGqlClient())
	defer cleanup(t, client)

	contents := "Test contents"
	unicode := true
	results, err := client.Send([]SendSmsInput{
		{
			Content: contents,
			Recipient: []Recipient{
				{
					PhoneNumber: "0123456789",
				},
			},
			Metadata: &SmsMetadata{
				Unicode: &unicode,
			},
			Save: true,
		},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var getQuery struct {
		SmsRequests []struct {
			ID string `json:"id"`
		} `graphql:"sms_request(where: $where)"`
	}

	getVariables := map[string]interface{}{
		"where": sms_request_bool_exp{
			"id": map[string]interface{}{
				"_eq": results.Responses[0].RequestID,
			},
		},
	}
	err = client.client.Query(context.TODO(), &getQuery, getVariables)
	if err != nil {
		t.Fatal(err)
	}
	if len(getQuery.SmsRequests) != 1 {
		t.Fatalf("expected 1 request, got: %d", len(getQuery.SmsRequests))
	}
}
