package docker_test

import (
	"fmt"
	"roob.re/reroller/registry/docker"
	"testing"
)

func Test_Dockerhub(t *testing.T) {
	infofunc := docker.DockerLikeImageInfo(docker.BaseUrl, docker.AuthUrl)

	info, err := infofunc("bitnami/nginx", "latest")
	if err != nil {
		t.Fatalf("error getting info from dockerhub: %v", err)
	}

	fmt.Println(info)
}
