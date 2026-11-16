package portal

import (
	"context"
	"net/http"
	"strings"
)

// DefaultAPI is the portal's functions URL on GarageAI's own domain; GARAGEAI_PORTAL_API overrides
// it (tests, staging). Door talks to this name so that the portal can move to another host without
// every installed Door having to change: the name is re-pointed, the garages notice nothing.
const DefaultAPI = "https://portal.garageai.eu/functions/v1"

// LegacyAPI is where the portal's functions run today, and the address Door used until v1.7.0. It
// is in the configuration of every garage connected before that (and in the link the portal hands
// out), so it is mapped to DefaultAPI when read; it is also the fallback while both exist.
const LegacyAPI = "https://sjopaiquzqotymqmrcke.supabase.co/functions/v1"

// Variables so that tests can point both at local servers.
var ownAPI, legacyAPI = DefaultAPI, LegacyAPI

// ownAddress maps a saved legacy address to the same path on GarageAI's own domain. Any other
// address (a test server, an operator's own --register-url) is returned as it is.
func ownAddress(url string) string {
	if rest, ok := strings.CutPrefix(url, legacyAPI); ok && (rest == "" || rest[0] == '/') {
		return ownAPI + rest
	}
	return url
}

// send does the request built by mk(url). When url is on GarageAI's own domain and the request did
// not reach the portal (no connection, or 502/503/504 from the proxy in front of it), it is sent
// once more to the same path on LegacyAPI: Door keeps working when the own name is not there yet
// or is down. A request that timed out or was cancelled is not sent again.
func send(ctx context.Context, client *http.Client, url string, mk func(url string) (*http.Request, error)) (*http.Response, error) {
	req, err := mk(url)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	rest, own := strings.CutPrefix(url, ownAPI)
	if !own || ctx.Err() != nil {
		return resp, err
	}
	if err == nil {
		switch resp.StatusCode {
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			resp.Body.Close()
		default:
			return resp, nil
		}
	}
	req, mkErr := mk(legacyAPI + rest)
	if mkErr != nil {
		return nil, mkErr
	}
	return client.Do(req)
}
