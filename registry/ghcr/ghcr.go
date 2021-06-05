package ghcr

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

const (
	Registry = "ghcr.io"
	BaseUrl  = "https://ghcr.io/v2"
)

func GHCRLikeImageInfo(baseUrl, user, password string) func(image, tag string) ([]string, error) {
	return func(image, tag string) ([]string, error) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(strings.Trim(baseUrl, "/")+"/%s/manifests/%s", image, tag), nil)
		if err != nil {
			log.Errorf("error building request: %v", err)
		}
		log.Debugf("querying ghcr as user '%s'", user)
		req.SetBasicAuth(user, password)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("ghcr API returned %d", resp.StatusCode)
		}

		var partialresponse struct {
			Config struct {
				Digest string `json:"digest"`
			} `json:"config"`
		}
		err = json.NewDecoder(resp.Body).Decode(&partialresponse)
		if err != nil {
			return nil, err
		}

		return []string{partialresponse.Config.Digest}, nil
	}
}
