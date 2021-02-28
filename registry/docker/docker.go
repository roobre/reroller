package docker

import (
	"encoding/json"
	"fmt"
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
			return nil, fmt.Errorf("docker API did not return any image")
		}

		digests := make([]string, len(partialresponse.Images))
		for i := range partialresponse.Images {
			digests[i] = partialresponse.Images[i].Digest
		}

		return digests, nil
	}
}
