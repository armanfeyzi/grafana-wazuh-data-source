package plugin

import (
	"testing"

	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

func TestValidateQueryRuleLevels(t *testing.T) {
	t.Parallel()

	min := 10
	max := 5
	err := validateQuery(models.Query{
		DataType: models.DataTypeAlerts,
		Format:   models.QueryFormatTable,
		Filters: models.QueryFilters{
			RuleLevelMin: &min,
			RuleLevelMax: &max,
		},
	})
	if err == nil {
		t.Fatal("expected rule level validation error")
	}
}

func TestValidateQueryAgentFormat(t *testing.T) {
	t.Parallel()

	err := validateQuery(models.Query{
		DataType: models.DataTypeAgents,
		Format:   models.QueryFormatTimeSeries,
	})
	if err == nil {
		t.Fatal("expected format validation error for agents")
	}
}

func TestValidateQueryUnimplementedType(t *testing.T) {
	t.Parallel()

	err := validateQuery(models.Query{DataType: models.DataTypeFIM})
	if err == nil {
		t.Fatal("expected unimplemented data type error")
	}
}
