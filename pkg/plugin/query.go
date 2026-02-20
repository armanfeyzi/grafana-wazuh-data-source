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
	case models.DataTypeVulnerabilities:
		return d.queryVulnerabilities(ctx, refID, qm, gq)
	case models.DataTypeFIM:
		return d.queryFIM(ctx, refID, qm, gq)
	case models.DataTypeSCA:
		return d.querySCA(ctx, refID, qm, gq)
	default:
		return backend.DataResponse{}, fmt.Errorf("data type %q is not implemented yet", qm.DataType)
	}
}

func (d *Datasource) queryAlerts(ctx context.Context, refID string, qm models.Query, gq backend.DataQuery) (backend.DataResponse, error) {
	params := indexer.AlertQueryParamsFrom(gq, qm)
	index := d.settings.AlertsIndexPattern()
	format := defaultFormat(qm.Format, models.QueryFormatTimeSeries)

	body, err := buildAlertsQueryBody(format, params)
	if err != nil {
		return backend.DataResponse{}, err
	}

	raw, err := d.indexer.Search(ctx, index, body)
	if err != nil {
		return backend.DataResponse{}, err
	}

	return parseAlertsResponse(format, raw, refID)
}

func buildAlertsQueryBody(format models.QueryFormat, params indexer.QueryParams) ([]byte, error) {
	switch format {
	case models.QueryFormatTimeSeries:
		return indexer.BuildAlertsTimeSeriesQuery(params)
	case models.QueryFormatTable:
		return indexer.BuildAlertsTableQuery(params)
	case models.QueryFormatStat:
		return indexer.BuildAlertsStatQuery(params)
	default:
		return nil, fmt.Errorf("unsupported alert format %q", format)
	}
}

func parseAlertsResponse(format models.QueryFormat, raw []byte, refID string) (backend.DataResponse, error) {
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

func (d *Datasource) queryVulnerabilities(ctx context.Context, refID string, qm models.Query, gq backend.DataQuery) (backend.DataResponse, error) {
	params := indexer.AlertQueryParamsFrom(gq, qm)
	index := d.settings.VulnerabilitiesIndexPattern()
	format := defaultFormat(qm.Format, models.QueryFormatTable)

	var (
		body []byte
		err  error
	)

	switch format {
	case models.QueryFormatTable:
		body, err = indexer.BuildVulnerabilitiesTableQuery(params)
	case models.QueryFormatStat:
		body, err = indexer.BuildVulnerabilitiesStatQuery(params)
	case models.QueryFormatTimeSeries:
		body, err = indexer.BuildVulnerabilitiesTimeSeriesQuery(params)
	default:
		return backend.DataResponse{}, fmt.Errorf("unsupported vulnerability format %q", format)
	}
	if err != nil {
		return backend.DataResponse{}, err
	}

	raw, err := d.indexer.Search(ctx, index, body)
	if err != nil {
		return backend.DataResponse{}, err
	}

	switch format {
	case models.QueryFormatTable:
		frames, err := indexer.ParseVulnerabilitiesTableFrames(raw, refID)
		if err != nil {
			return backend.DataResponse{}, err
		}
		return backend.DataResponse{Frames: frames}, nil
	case models.QueryFormatStat:
		frame, err := indexer.ParseVulnerabilitiesStatFrame(raw, refID)
		if err != nil {
			return backend.DataResponse{}, err
		}
		return backend.DataResponse{Frames: []*data.Frame{frame}}, nil
	case models.QueryFormatTimeSeries:
		frame, err := indexer.ParseVulnerabilitiesTimeSeriesFrame(raw, refID)
		if err != nil {
			return backend.DataResponse{}, err
		}
		return backend.DataResponse{Frames: []*data.Frame{frame}}, nil
	default:
		return backend.DataResponse{}, fmt.Errorf("unsupported vulnerability format %q", format)
	}
}

func (d *Datasource) queryFIM(ctx context.Context, refID string, qm models.Query, gq backend.DataQuery) (backend.DataResponse, error) {
	params := indexer.AlertQueryParamsFrom(gq, qm)
	index := d.settings.AlertsIndexPattern()
	format := defaultFormat(qm.Format, models.QueryFormatTimeSeries)

	var (
		body []byte
		err  error
	)

	switch format {
	case models.QueryFormatTimeSeries:
		body, err = indexer.BuildFIMTimeSeriesQuery(params)
	case models.QueryFormatTable:
		body, err = indexer.BuildFIMTableQuery(params)
	case models.QueryFormatStat:
		body, err = indexer.BuildFIMStatQuery(params)
	default:
		return backend.DataResponse{}, fmt.Errorf("unsupported FIM format %q", format)
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
		frame, err := indexer.ParseFIMTimeSeriesFrame(raw, refID)
		if err != nil {
			return backend.DataResponse{}, err
		}
		return backend.DataResponse{Frames: []*data.Frame{frame}}, nil
	case models.QueryFormatTable:
		frames, err := indexer.ParseFIMTableFrames(raw, refID)
		if err != nil {
			return backend.DataResponse{}, err
		}
		return backend.DataResponse{Frames: frames}, nil
	case models.QueryFormatStat:
		frame, err := indexer.ParseFIMStatFrame(raw, refID)
		if err != nil {
			return backend.DataResponse{}, err
		}
		return backend.DataResponse{Frames: []*data.Frame{frame}}, nil
	default:
		return backend.DataResponse{}, fmt.Errorf("unsupported FIM format %q", format)
	}
}

func (d *Datasource) querySCA(ctx context.Context, refID string, qm models.Query, gq backend.DataQuery) (backend.DataResponse, error) {
	format := defaultFormat(qm.Format, models.QueryFormatTable)

	switch format {
	case models.QueryFormatTable:
		return d.querySCALive(ctx, refID, qm)
	case models.QueryFormatTimeSeries, models.QueryFormatStat:
		return d.querySCAHistory(ctx, refID, qm, gq, format)
	default:
		return backend.DataResponse{}, fmt.Errorf("unsupported SCA format %q", format)
	}
}

func (d *Datasource) querySCALive(ctx context.Context, refID string, qm models.Query) (backend.DataResponse, error) {
	raw, err := d.wazuhAPI.ListSCAForAgents(ctx, qm.Filters.AgentNames, qm.Limit)
	if err != nil {
		return backend.DataResponse{}, err
	}

	frame, err := wazuhapi.ParseSCALiveTableFrame(raw, refID)
	if err != nil {
		return backend.DataResponse{}, err
	}

	return backend.DataResponse{Frames: []*data.Frame{frame}}, nil
}

func (d *Datasource) querySCAHistory(ctx context.Context, refID string, qm models.Query, gq backend.DataQuery, format models.QueryFormat) (backend.DataResponse, error) {
	params := indexer.AlertQueryParamsFrom(gq, qm)
	index := d.settings.AlertsIndexPattern()

	var (
		body []byte
		err  error
	)

	switch format {
	case models.QueryFormatTimeSeries:
		body, err = indexer.BuildSCAHistoryTimeSeriesQuery(params)
	case models.QueryFormatStat:
		body, err = indexer.BuildSCAHistoryStatQuery(params)
	default:
		return backend.DataResponse{}, fmt.Errorf("unsupported SCA history format %q", format)
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
		frame, err := indexer.ParseSCAHistoryTimeSeriesFrame(raw, refID)
		if err != nil {
			return backend.DataResponse{}, err
		}
		return backend.DataResponse{Frames: []*data.Frame{frame}}, nil
	case models.QueryFormatStat:
		frame, err := indexer.ParseSCAHistoryStatFrame(raw, refID)
		if err != nil {
			return backend.DataResponse{}, err
		}
		return backend.DataResponse{Frames: []*data.Frame{frame}}, nil
	default:
		return backend.DataResponse{}, fmt.Errorf("unsupported SCA history format %q", format)
	}
}

func defaultFormat(format, fallback models.QueryFormat) models.QueryFormat {
	if format == "" {
		return fallback
	}
	return format
}
