package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const dockerBaseURL = "https://registry.hub.docker.com/v2/repositories"

func dockerImageInfoFunc(image, tag string) (string, error) {
	resp, err := http.Get(fmt.Sprintf(dockerBaseURL+"/%s/tags/%s", image, tag))
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("docker API returned %d", resp.StatusCode)
	}

	var partialresponse struct {
		Images []struct {
			Digest string `json:"digest"`
		} `json:"images"`
	}
	err = json.NewDecoder(resp.Body).Decode(&partialresponse)
	if err != nil {
		return "", err
	}

	if len(partialresponse.Images) < 1 {
		return "", fmt.Errorf("docker API did not return %d", resp.StatusCode)
	}

	return partialresponse.Images[0].Digest, nil
}

var DockerImageInfoFunc = ImageInfoFunc(dockerImageInfoFunc)
