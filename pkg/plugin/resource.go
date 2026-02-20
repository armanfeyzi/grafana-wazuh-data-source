package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/wazuhapi"
)

type agentOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	if req.Method != http.MethodGet {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusMethodNotAllowed,
			Body:   []byte(`{"message":"method not allowed"}`),
		})
	}

	switch req.Path {
	case "agents":
		return d.callResourceAgents(ctx, sender)
	default:
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusNotFound,
			Body:   []byte(fmt.Sprintf(`{"message":"unknown resource %q"}`, req.Path)),
		})
	}
}

func (d *Datasource) callResourceAgents(ctx context.Context, sender backend.CallResourceResponseSender) error {
	raw, err := d.wazuhAPI.ListAgents(ctx, 500)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadGateway,
			Body:   []byte(fmt.Sprintf(`{"message":%q}`, err.Error())),
		})
	}

	options, err := wazuhapi.ParseAgentOptions(raw)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf(`{"message":%q}`, err.Error())),
		})
	}

	body, err := json.Marshal(options)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(`{"message":"failed to encode agent list"}`),
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   body,
	})
}
