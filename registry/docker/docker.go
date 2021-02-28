package docker

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

func DockerLikeImageInfo(baseUrl string) func(image, tag string) ([]string, error) {
	return func(image, tag string) ([]string, error) {
		resp, err := http.Get(fmt.Sprintf(strings.Trim(baseUrl, "/")+"/%s/tags/%s", image, tag))
		if err != nil {
			return nil, err
		}

		if resp.StatusCode >= 400 {
			log.Debug(fmt.Sprintf(baseUrl+"/%s/tags/%s", image, tag))
			log.Debug(ioutil.ReadAll(resp.Body))
			return nil, fmt.Errorf("docker API returned %d", resp.StatusCode)
		}

		var partialresponse struct {
			Images []struct {
				Digest string `json:"digest"`
			} `json:"images"`
		}
		err = json.NewDecoder(resp.Body).Decode(&partialresponse)
		if err != nil {
			return nil, err
		}

		if len(partialresponse.Images) < 1 {
			return nil, fmt.Errorf("docker API returned %d", resp.StatusCode)
		}

		digests := make([]string, len(partialresponse.Images))
		for i := range partialresponse.Images {
			digests[i] = partialresponse.Images[i].Digest
		}

		return digests, nil
	}
}
