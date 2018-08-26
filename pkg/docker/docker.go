package docker

import (
	"net"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/overlordtm/docker-dns/pkg/config"
	"github.com/sirupsen/logrus"
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
	logrus.WithFields(logrus.Fields{"containerID": id, "isAlias": isAlias}).Debug("FindContainerIP")

	c, err = dc.dockerClient.InspectContainer(id)

	if err != nil {
		logrus.WithFields(logrus.Fields{"isAlias": isAlias, "id": id, "err": err}).Debug("Failed to find container by name or id")
		return
	}

	if c.NetworkSettings.IPAddress != "" {
		ip = net.ParseIP(c.NetworkSettings.IPAddress).To4()
	} else {
		for name, network := range c.NetworkSettings.Networks {
			if strings.Contains(strings.ToLower(name), "default") {
				ip = net.ParseIP(network.IPAddress).To4()
				break
			}
		}
	}

	logrus.WithField("dockerResponse", c.NetworkSettings).Debug("Docker response")

	return
}
