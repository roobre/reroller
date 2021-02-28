package ghcr

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

func GHCRLikeImageInfo(baseUrl, user, password string) func(image, tag string) ([]string, error) {
	return func(image, tag string) ([]string, error) {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(strings.Trim(baseUrl, "/")+"/%s/manifests/%s", image, tag), nil)
		if err != nil {
			log.Errorf("error building request: %v", err)
		}
		log.Debugf("querying ghcr as %s", user)
		req.SetBasicAuth(user, password)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode >= 400 {
			log.Debug(fmt.Sprintf(baseUrl+"/%s/tags/%s", image, tag))
			log.Debug(ioutil.ReadAll(resp.Body))
			return nil, fmt.Errorf("docker API returned %d", resp.StatusCode)
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
