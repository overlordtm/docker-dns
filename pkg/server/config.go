package server


type Config struct {
	TLD            string
	TTL            uint32
	Listen         string
	DockerEndpoint string
	Aliases        map[string]string
}