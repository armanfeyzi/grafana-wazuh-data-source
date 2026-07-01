package indexer

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/armanfeyzi/grafana-wazuh-data-source-plugin/pkg/models"
)

// classifyHTTPError maps indexer HTTP status codes to typed WazuhErrors.
func classifyHTTPError(status int) *models.WazuhError {
	switch status {
	case http.StatusUnauthorized:
		return models.NewWazuhError(models.ErrAuth,
			"Wazuh indexer: authentication failed — check indexer username and password", nil)
	case http.StatusForbidden:
		return models.NewWazuhError(models.ErrForbidden,
			"Wazuh indexer: permission denied — check indexer RBAC (read access to wazuh-* indices)", nil)
	case http.StatusNotFound:
		return models.NewWazuhError(models.ErrIndexMissing,
			"Wazuh indexer: index not found — Wazuh may not have generated this data yet", nil)
	default:
		return models.NewWazuhError(models.ErrBadResponse,
			fmt.Sprintf("Wazuh indexer returned an unexpected response (HTTP %d)", status), nil)
	}
}

// classifyNetworkError maps transport errors to ErrTimeout or ErrUnreachable.
func classifyNetworkError(err error) *models.WazuhError {
	if errors.Is(err, context.DeadlineExceeded) {
		return models.NewWazuhError(models.ErrTimeout,
			"Wazuh indexer: query timed out — try a shorter time range", err)
	}
	return models.NewWazuhError(models.ErrUnreachable,
		"Wazuh indexer: cannot connect — check the indexer URL and that the service is running", err)
}
