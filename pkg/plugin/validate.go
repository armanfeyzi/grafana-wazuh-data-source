package plugin

import (
	"fmt"

	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

func validateQuery(qm models.Query) error {
	if qm.DataType == "" {
		return fmt.Errorf("dataType is required")
	}

	switch qm.DataType {
	case models.DataTypeAlerts,
		models.DataTypeAgents,
		models.DataTypeVulnerabilities,
		models.DataTypeFIM,
		models.DataTypeSCA:
	default:
		return fmt.Errorf("unknown data type %q", qm.DataType)
	}

	if err := validateFormat(qm); err != nil {
		return err
	}

	if err := validateFilters(qm.Filters); err != nil {
		return err
	}

	if qm.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}

	return nil
}

func validateFormat(qm models.Query) error {
	format := qm.Format
	if format == "" {
		return nil
	}

	switch qm.DataType {
	case models.DataTypeAlerts, models.DataTypeFIM:
		switch format {
		case models.QueryFormatTimeSeries, models.QueryFormatTable, models.QueryFormatStat:
			return nil
		default:
			return fmt.Errorf("unsupported format %q", format)
		}
	case models.DataTypeVulnerabilities:
		switch format {
		case models.QueryFormatTable, models.QueryFormatStat, models.QueryFormatTimeSeries:
			return nil
		default:
			return fmt.Errorf("unsupported format %q", format)
		}
	case models.DataTypeAgents:
		if format != models.QueryFormatTable {
			return fmt.Errorf("agent status only supports table format")
		}
	case models.DataTypeSCA:
		switch format {
		case models.QueryFormatTable, models.QueryFormatTimeSeries, models.QueryFormatStat:
			return nil
		default:
			return fmt.Errorf("unsupported format %q", format)
		}
	}

	return nil
}

func validateFilters(f models.QueryFilters) error {
	if f.RuleLevelMin != nil && (*f.RuleLevelMin < 0 || *f.RuleLevelMin > 15) {
		return fmt.Errorf("rule level min must be between 0 and 15")
	}
	if f.RuleLevelMax != nil && (*f.RuleLevelMax < 0 || *f.RuleLevelMax > 15) {
		return fmt.Errorf("rule level max must be between 0 and 15")
	}
	if f.RuleLevelMin != nil && f.RuleLevelMax != nil && *f.RuleLevelMin > *f.RuleLevelMax {
		return fmt.Errorf("rule level min cannot be greater than max")
	}
	return nil
}
