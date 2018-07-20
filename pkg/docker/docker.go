package docker

import (
	"net"
	"github.com/fsouza/go-dockerclient"
	"github.com/sirupsen/logrus"
	"github.com/overlordtm/docker-dns/pkg/config"
)

type DockerFinder interface {
	FindContainerIP(id string) (ip net.IP, err error)
}

type DockerFinderImpl struct {
	dockerClient *docker.Client
	config       *config.Config
}

func New(config *config.Config) (d *DockerFinderImpl, err error) {
	dockerClient, err := docker.NewClient(config.DockerEndpoint)
	if err != nil {
		return
	}

	d = &DockerFinderImpl{
		dockerClient: dockerClient,
		config:       config,
	}

	return
}

func (dc *DockerFinderImpl) FindContainerIP(id string) (ip net.IP, err error) {

	var c *docker.Container
	isAlias := false
	alias, isAlias := dc.config.Aliases[id]

	if isAlias {
		id = alias
	}

	c, err = dc.dockerClient.InspectContainer(id)

	if err != nil {
		logrus.WithFields(logrus.Fields{"isAlias": isAlias, "id": id, "err": err}).Debug("Failed to find container by name or id")
		return
	}

	ip = net.ParseIP(c.NetworkSettings.IPAddress).To4()

	return
}
