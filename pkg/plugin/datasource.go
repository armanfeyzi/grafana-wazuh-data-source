package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/httpclient"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/indexer"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/wazuhapi"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
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

func (d *Datasource) query(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	var qm models.Query

	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err))
	}

	if qm.DataType == "" {
		return backend.ErrDataResponse(backend.StatusBadRequest, "dataType is required")
	}

	if err := validateQuery(qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
	}

	resp, err := d.executeQuery(ctx, query.RefID, qm, query)
	if err != nil {
		return wazuhErrToDataResponse(err)
	}

	return resp
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
		return healthError("Wazuh manager API: " + models.UserMessage(err)), nil
	}

	if err := d.indexer.Ping(ctx); err != nil {
		return healthError("Wazuh indexer: " + models.UserMessage(err)), nil
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

// wazuhErrToDataResponse converts a WazuhError (or plain error) into a Grafana
// DataResponse with an appropriate HTTP status code and user-readable message.
func wazuhErrToDataResponse(err error) backend.DataResponse {
	we, ok := models.AsWazuhError(err)
	if !ok {
		return backend.ErrDataResponse(backend.StatusInternal, err.Error())
	}

	switch we.Code {
	case models.ErrAuth:
		return backend.ErrDataResponse(http.StatusUnauthorized, we.Message)
	case models.ErrForbidden:
		return backend.ErrDataResponse(http.StatusForbidden, we.Message)
	case models.ErrIndexMissing:
		return backend.ErrDataResponse(http.StatusNotFound, we.Message)
	case models.ErrTimeout:
		return backend.ErrDataResponse(http.StatusGatewayTimeout, we.Message)
	case models.ErrUnreachable:
		return backend.ErrDataResponse(http.StatusBadGateway, we.Message)
	default:
		return backend.ErrDataResponse(backend.StatusInternal, we.Message)
	}
}
