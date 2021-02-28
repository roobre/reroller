package docker

const dockerBaseURL = "https://registry.hub.docker.com/v2/repositories"

func ImageInfo(image, tag string) ([]string, error) {
	return DockerLikeImageInfo(dockerBaseURL, image, tag)
}
