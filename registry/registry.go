package registry

import (
	"errors"
	"roob.re/reroller/registry/docker"
	"strings"
)

func ImageDigests(image string) ([]string, error) {
	id := ParseImage(image)

	infoFunc := ImageInfoFunc(func(image, tag string) ([]string, error) {
		return nil, errors.New("unknown registry for image " + image)
	})

	switch id.Registry {
	case "docker.io":
		infoFunc = docker.ImageInfo
	}

	return infoFunc(id.Name, id.Tag)
}

// ImageInfoFunc is able to provide the latest SHA of an image given its name and tag
type ImageInfoFunc func(image, tag string) ([]string, error)

type ImageDescriptor struct {
	Registry string
	Name     string
	Tag      string
}

const defaultRegistry = "docker.io"
const defaultTag = "latest"

func ParseImage(image string) ImageDescriptor {
	d := ImageDescriptor{}

	pieces := strings.Split(image, "/")

	if len(pieces) >= 3 {
		// Has registry
		d.Registry = pieces[0]
		pieces = pieces[1:]
	} else {
		d.Registry = defaultRegistry
	}

	if len(pieces) < 2 {
		pieces[0] = "library/" + pieces[0]
	}

	repotag := strings.Split(strings.Join(pieces, "/"), ":")
	d.Name = repotag[0]
	if len(repotag) >= 2 {
		// Has tag
		d.Tag = repotag[1]
	} else {
		d.Tag = defaultTag
	}

	return d
}
