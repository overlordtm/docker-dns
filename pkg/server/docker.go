package server

import "net"

func findContainerIP(id string) (ip net.IP, err error) {

	var c *docker.Container
	alias, isAlias := conf.Aliases[id]

	if isAlias {
		id = alias
	}

	c, err = dockerClient.InspectContainer(id)

	if err != nil {
		logrus.WithFields(logrus.Fields{"isAlias": isAlias, "id": id, "err": err}).Debug("Failed to find container by name or id")
		return
	}

	ip = net.ParseIP(c.NetworkSettings.IPAddress).To4()

	return
}