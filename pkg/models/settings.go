package models

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type PluginSettings struct {
	ManagerURL      string                `json:"managerUrl"`
	IndexerURL      string                `json:"indexerUrl"`
	Username        string                `json:"username"`
	IndexerUsername string                `json:"indexerUsername"`
	TlsSkipVerify   bool                  `json:"tlsSkipVerify"`
	IndexPrefix     string                `json:"indexPrefix"`
	Secrets         *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	Password        string
	IndexerPassword string
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}
	if err := json.Unmarshal(source.JSONData, &settings); err != nil {
		return nil, fmt.Errorf("unmarshal plugin settings: %w", err)
	}

	settings.Secrets = loadSecretPluginSettings(source.DecryptedSecureJSONData)

	return &settings, nil
}

func (s *PluginSettings) IndexerUser() string {
	if s.IndexerUsername != "" {
		return s.IndexerUsername
	}
	return s.Username
}

func (s *PluginSettings) IndexerPass() string {
	if s.Secrets == nil {
		return ""
	}
	if s.Secrets.IndexerPassword != "" {
		return s.Secrets.IndexerPassword
	}
	return s.Secrets.Password
}

func (s *PluginSettings) AlertsIndexPattern() string {
	if s.IndexPrefix == "" {
		return "wazuh-alerts-*"
	}
	if strings.HasSuffix(s.IndexPrefix, "*") {
		return s.IndexPrefix
	}
	return s.IndexPrefix + "*"
}

func loadSecretPluginSettings(source map[string]string) *SecretPluginSettings {
	return &SecretPluginSettings{
		Password:        source["password"],
		IndexerPassword: source["indexerPassword"],
	}
}
