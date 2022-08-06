package sms

import (
	"context"
	"time"

	"github.com/hasura/go-graphql-client"
)

// Client a high level sms client that communicates the backend through GraphQL client
type Client struct {
	client *graphql.Client
}

// New constructs a sms client
func New(client *graphql.Client) *Client {
	return &Client{
		client: client,
	}
}

// Send a sms request
func (c *Client) Send(inputs []SendSmsInput, variables map[string]string) (*SendSmsOutput, error) {
	if len(inputs) == 0 {
		return &SendSmsOutput{}, nil
	}

	for i, input := range inputs {
		if input.SendAfter.IsZero() {
			inputs[i].SendAfter = time.Now()
		}
	}
	var mutation struct {
		SendSms SendSmsOutput `graphql:"sendSMS(data: $data, variables: $variables)"`
	}

	inputVariables := map[string]interface{}{
		"data":      inputs,
		"variables": json(variables),
	}

	err := c.client.Mutate(context.Background(), &mutation, inputVariables, graphql.OperationName("SendSMSs"))
	if err != nil {
		return nil, err
	}

	return &mutation.SendSms, nil
}

// CancelSms cancel and delete sms requests
func (c *Client) CancelSms(where map[string]interface{}) (int, error) {

	var mutation struct {
		DeleteSMSs struct {
			AffectedRows int `graphql:"affected_rows"`
		} `graphql:"delete_sms_request(where: $where)"`
	}

	variables := map[string]interface{}{
		"where": sms_request_bool_exp(where),
	}

	err := c.client.Mutate(context.Background(), &mutation, variables, graphql.OperationName("DeleteSmsRequests"))
	if err != nil {
		return 0, err
	}

	return mutation.DeleteSMSs.AffectedRows, nil
}
