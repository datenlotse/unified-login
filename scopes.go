package unified_login

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Scope struct {
	Scope       string `json:"scope"`
	Description string `json:"description"`
}

type scopeSyncRequestBody struct {
	Secret  string  `json:"secret"`
	Scopes  []Scope `json:"scopes"`
	OwnerId string  `json:"ownerId"`
}

func SyncScopes(ctx context.Context, host string, clientSecret string, systemUserId string, scopes []Scope) error {
	body, err := json.Marshal(scopeSyncRequestBody{
		Secret:  clientSecret,
		Scopes:  scopes,
		OwnerId: systemUserId,
	})
	if err != nil {
		return fmt.Errorf("Error marshaling json data: %w", err)
	}

	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/apps/scopes", host), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("Error creating http request object: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error during http request: %w", err)
	}
	if res.StatusCode >= 300 || res.StatusCode < 200 {
		return fmt.Errorf("Recieved non 200 status code: %s", res.Status)
	}
	defer res.Body.Close()

	return nil
}
