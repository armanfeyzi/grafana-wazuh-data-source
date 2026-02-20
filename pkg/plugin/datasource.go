package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

func NewDatasource(_ context.Context, _ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &Datasource{}, nil
}

type Datasource struct{}

func (d *Datasource) Dispose() {}

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		response.Responses[q.RefID] = d.query(ctx, q)
	}

	return response, nil
}

func (d *Datasource) query(_ context.Context, query backend.DataQuery) backend.DataResponse {
	var qm models.Query

	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err))
	}

	if qm.DataType == "" {
		return backend.ErrDataResponse(backend.StatusBadRequest, "dataType is required")
	}

	// Placeholder response until Phase 2 query engine is implemented.
	frame := data.NewFrame("response")
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{query.TimeRange.From, query.TimeRange.To}),
		data.NewField("values", nil, []int64{10, 20}),
	)

	return backend.DataResponse{Frames: []*data.Frame{frame}}
}

func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Unable to load settings",
		}, nil
	}

	switch {
	case config.ManagerURL == "":
		return healthError("Manager URL is required"), nil
	case config.IndexerURL == "":
		return healthError("Indexer URL is required"), nil
	case config.Username == "":
		return healthError("Username is required"), nil
	case config.Secrets == nil || config.Secrets.Password == "":
		return healthError("Password is required"), nil
	}

	// Connectivity checks land in Phase 1.
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Configuration looks valid",
	}, nil
}

func healthError(message string) *backend.CheckHealthResult {
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusError,
		Message: message,
	}
}
