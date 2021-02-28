package docker

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Many registries implement docker-compatible API. This function will work for them, provided their base URL
func DockerLikeImageInfo(baseurl, image, tag string) ([]string, error) {
	resp, err := http.Get(fmt.Sprintf(baseurl+"/%s/tags/%s", image, tag))
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
		return nil, fmt.Errorf("docker API returned %d", resp.StatusCode)
	}

	digests := make([]string, len(partialresponse.Images))
	for i := range partialresponse.Images {
		digests[i] = partialresponse.Images[i].Digest
	}

	return digests, nil
}
