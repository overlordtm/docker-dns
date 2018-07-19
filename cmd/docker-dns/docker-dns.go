package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/coreos/go-systemd/daemon"
	"docker-dns/pkg/config"
	"docker-dns/pkg/server"
)

const (
	DEFAULT_CONFIG_PATH = "/etc/docker-dns/config.json"
	DEFAULT_DOCKER_ADDR = "unix:///var/run/docker.sock"
	DEFAULT_LISTEN_ADDR = "127.0.0.1:8053"
)

func main() {
	var (
		err            error
		listenAddr     string
		dockerAddr     string
		ttl            uint64
		tld            string
		logLevel       string
		configFilePath string
		systemd        bool
		createConfig   bool

		conf *config.Config
	)

	flag.StringVar(&listenAddr, "listen", "", "Listen address")
	flag.StringVar(&dockerAddr, "docker", "", "Address for docker client (HTTP or Unix)")
	flag.StringVar(&tld, "tld", "", "TLD to serve")
	flag.StringVar(&logLevel, "loglevel", "info", "Logrus loglevel")
	flag.StringVar(&configFilePath, "config", "", "Path to config file")
	flag.Uint64Var(&ttl, "ttl", 0, "Default TTL")
	flag.BoolVar(&systemd, "systemd", false, "Start as systemd service")

	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	// logging setup
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.WithField("err", err).Fatal("Failed to parse log level")
	}
	logrus.SetLevel(lvl)

	if configFilePath != "" && createConfig == false {
		conf, err = config.ReadConfig(configFilePath)
		if err != nil {
			logrus.WithField("err", err).Fatal("Failed to load config")
		}
	} else {
		conf = new(config.Config)
	}

	logrus.WithField("config", conf).Info("Starting server")

	srv := server.New(conf)
	if err != nil {
		logrus.WithField("err", err).Fatal("Failed to create docker client")
	}

	go srv.Handle()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	if systemd == true {
		daemon.SdNotify(false, "READY=1")
	}

forever:
	for {
		select {
		case _ = <-sig:
			break forever
		}
	}
}
