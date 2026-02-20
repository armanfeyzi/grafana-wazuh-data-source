package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type PluginSettings struct {
	ManagerURL    string                `json:"managerUrl"`
	IndexerURL    string                `json:"indexerUrl"`
	Username      string                `json:"username"`
	TlsSkipVerify bool                  `json:"tlsSkipVerify"`
	IndexPrefix   string                `json:"indexPrefix"`
	Secrets       *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	Password string
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}
	if err := json.Unmarshal(source.JSONData, &settings); err != nil {
		return nil, fmt.Errorf("unmarshal plugin settings: %w", err)
	}

	settings.Secrets = loadSecretPluginSettings(source.DecryptedSecureJSONData)

	return &settings, nil
}

func loadSecretPluginSettings(source map[string]string) *SecretPluginSettings {
	return &SecretPluginSettings{
		Password: source["password"],
	}
}
