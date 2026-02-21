package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/indexer"
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
	case "namespaces":
		return d.callResourceNamespaces(ctx, sender)
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

func (d *Datasource) callResourceNamespaces(ctx context.Context, sender backend.CallResourceResponseSender) error {
	query, err := indexer.BuildNamespacesQuery(200)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf(`{"message":%q}`, err.Error())),
		})
	}

	index := d.settings.AlertsIndexPattern()
	raw, err := d.indexer.Search(ctx, index, query)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadGateway,
			Body:   []byte(fmt.Sprintf(`{"message":%q}`, err.Error())),
		})
	}

	namespaces, err := indexer.ParseNamespacesResponse(raw)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf(`{"message":%q}`, err.Error())),
		})
	}

	// Convert to the same label/value format as the agents endpoint so the
	// frontend can reuse the same AgentOption type.
	options := make([]agentOption, 0, len(namespaces))
	for _, ns := range namespaces {
		options = append(options, agentOption{Label: ns, Value: ns})
	}

	body, err := json.Marshal(options)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(`{"message":"failed to encode namespace list"}`),
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   body,
	})
}
