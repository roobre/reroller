package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	Registry = "docker.io"
	BaseUrl  = "https://index.docker.io/v2"
	AuthUrl  = "https://auth.docker.io"
)

func DockerLikeImageInfo(baseUrl, authUrl string) func(image, tag string) ([]string, error) {
	baseUrl = strings.Trim(baseUrl, "/")
	authUrl = strings.Trim(authUrl, "/")

	return func(image, tag string) ([]string, error) {
		digests := make([]string, 0, 2)

		// TODO: Get authUrl from WWW-Authenticate header
		authResp, err := http.Get(fmt.Sprintf(authUrl+"/token?service=registry.docker.io&scope=repository:%s:pull", image))
		if err != nil {
			return nil, fmt.Errorf("retrieving dockerhub token: %w", err)
		}
		if authResp.StatusCode >= 400 {
			return nil, fmt.Errorf("dockerhub auth endpoint returned %d", authResp.StatusCode)
		}

		var authResponse struct {
			Token string `json:"token"`
		}

		err = json.NewDecoder(authResp.Body).Decode(&authResponse)
		if err != nil {
			return nil, fmt.Errorf("decoding response from dockerhub auth endpoint: %w", err)
		}

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(baseUrl+"/%s/manifests/%s", image, tag), nil)
		if err != nil {
			return nil, fmt.Errorf("building manifest request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+authResponse.Token)

		// Query the manifest list digest
		req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
		retDigests, err := dockerContentDigest(req)
		if err != nil {
			return nil, fmt.Errorf("getting manifest list from docker: %w", err)
		}

		digests = append(digests, retDigests...)

		// Query the v2 manifest digest
		req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
		retDigests, err = dockerContentDigest(req)
		if err != nil {
			return nil, fmt.Errorf("getting manifest list from docker: %w", err)
		}

		digests = append(digests, retDigests...)

		return digests, nil
	}
}

func dockerContentDigest(req *http.Request) ([]string, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying dockerhub tags endpoint: %w", err)
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	digest, found := resp.Header["Docker-Content-Digest"]
	if !found {
		return nil, fmt.Errorf("docker API did not return a content digest for this image")
	}

	return digest, nil
}
