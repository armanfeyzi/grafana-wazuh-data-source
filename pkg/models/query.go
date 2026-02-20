package models

type DataType string

const (
	DataTypeAlerts          DataType = "alerts"
	DataTypeVulnerabilities DataType = "vulnerabilities"
	DataTypeFIM             DataType = "fim"
	DataTypeSCA             DataType = "sca"
	DataTypeAgents          DataType = "agents"
)

type QueryFormat string

const (
	QueryFormatTimeSeries QueryFormat = "time_series"
	QueryFormatTable      QueryFormat = "table"
	QueryFormatStat       QueryFormat = "stat"
)

type QueryFilters struct {
	AgentNames   []string `json:"agentNames,omitempty"`
	RuleLevelMin *int     `json:"ruleLevelMin,omitempty"`
	RuleLevelMax *int     `json:"ruleLevelMax,omitempty"`
	Severity     []string `json:"severity,omitempty"`
}

type Query struct {
	DataType    DataType     `json:"dataType"`
	Format      QueryFormat  `json:"format"`
	Aggregation string       `json:"aggregation,omitempty"`
	GroupBy     string       `json:"groupBy,omitempty"`
	Filters     QueryFilters `json:"filters,omitempty"`
	Limit       int          `json:"limit,omitempty"`
}
