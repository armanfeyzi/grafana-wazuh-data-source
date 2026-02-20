package plugin

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/indexer"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/wazuhapi"
)

func (d *Datasource) executeQuery(ctx context.Context, refID string, qm models.Query, gq backend.DataQuery) (backend.DataResponse, error) {
	switch qm.DataType {
	case models.DataTypeAlerts:
		return d.queryAlerts(ctx, refID, qm, gq)
	case models.DataTypeAgents:
		return d.queryAgents(ctx, refID, qm)
	default:
		return backend.DataResponse{}, fmt.Errorf("data type %q is not implemented yet", qm.DataType)
	}
}

func (d *Datasource) queryAlerts(ctx context.Context, refID string, qm models.Query, gq backend.DataQuery) (backend.DataResponse, error) {
	params := indexer.AlertQueryParamsFrom(gq, qm)
	index := d.settings.AlertsIndexPattern()

	format := qm.Format
	if format == "" {
		format = models.QueryFormatTimeSeries
	}

	var (
		body []byte
		err  error
	)

	switch format {
	case models.QueryFormatTimeSeries:
		body, err = indexer.BuildAlertsTimeSeriesQuery(params)
	case models.QueryFormatTable:
		body, err = indexer.BuildAlertsTableQuery(params)
	case models.QueryFormatStat:
		body, err = indexer.BuildAlertsStatQuery(params)
	default:
		return backend.DataResponse{}, fmt.Errorf("unsupported alert format %q", format)
	}
	if err != nil {
		return backend.DataResponse{}, err
	}

	raw, err := d.indexer.Search(ctx, index, body)
	if err != nil {
		return backend.DataResponse{}, err
	}

	switch format {
	case models.QueryFormatTimeSeries:
		frame, err := indexer.ParseAlertsTimeSeriesFrame(raw, refID)
		if err != nil {
			return backend.DataResponse{}, err
		}
		return backend.DataResponse{Frames: []*data.Frame{frame}}, nil
	case models.QueryFormatTable:
		frames, err := indexer.ParseAlertsTableFrames(raw, refID)
		if err != nil {
			return backend.DataResponse{}, err
		}
		return backend.DataResponse{Frames: frames}, nil
	case models.QueryFormatStat:
		frame, err := indexer.ParseAlertsStatFrame(raw, refID)
		if err != nil {
			return backend.DataResponse{}, err
		}
		return backend.DataResponse{Frames: []*data.Frame{frame}}, nil
	default:
		return backend.DataResponse{}, fmt.Errorf("unsupported alert format %q", format)
	}
}

func (d *Datasource) queryAgents(ctx context.Context, refID string, qm models.Query) (backend.DataResponse, error) {
	raw, err := d.wazuhAPI.ListAgents(ctx, qm.Limit)
	if err != nil {
		return backend.DataResponse{}, err
	}

	frame, err := wazuhapi.ParseAgentsFrame(raw, refID)
	if err != nil {
		return backend.DataResponse{}, err
	}

	return backend.DataResponse{Frames: []*data.Frame{frame}}, nil
}
