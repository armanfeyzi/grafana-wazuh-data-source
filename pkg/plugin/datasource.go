package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/httpclient"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/indexer"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/wazuhapi"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	config, err := models.LoadPluginSettings(settings)
	if err != nil {
		return nil, err
	}

	httpClient := httpclient.New(config.TlsSkipVerify)

	return &Datasource{
		settings: config,
		wazuhAPI: wazuhapi.NewClient(config, httpClient),
		indexer:  indexer.NewClient(config, httpClient),
	}, nil
}

type Datasource struct {
	settings *models.PluginSettings
	wazuhAPI *wazuhapi.Client
	indexer  *indexer.Client
}

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

func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		return healthError("Unable to load settings"), nil
	}

	if err := validateSettings(config); err != nil {
		return healthError(err.Error()), nil
	}

	if err := wazuhapi.ValidateURL(config.ManagerURL); err != nil {
		return healthError(err.Error()), nil
	}
	if err := indexer.ValidateURL(config.IndexerURL); err != nil {
		return healthError(err.Error()), nil
	}

	if err := d.wazuhAPI.Ping(ctx); err != nil {
		return healthError(fmt.Sprintf("Wazuh manager API: %v", err)), nil
	}

	if err := d.indexer.Ping(ctx); err != nil {
		return healthError(fmt.Sprintf("Wazuh indexer: %v", err)), nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Connected to Wazuh manager API and indexer",
	}, nil
}

func validateSettings(config *models.PluginSettings) error {
	switch {
	case config.Username == "":
		return fmt.Errorf("Username is required")
	case config.Secrets == nil || config.Secrets.Password == "":
		return fmt.Errorf("Password is required")
	case config.IndexerUser() == "":
		return fmt.Errorf("Indexer username is required")
	case config.IndexerPass() == "":
		return fmt.Errorf("Indexer password is required")
	default:
		return nil
	}
}

func healthError(message string) *backend.CheckHealthResult {
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusError,
		Message: message,
	}
}
