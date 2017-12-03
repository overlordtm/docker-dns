package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/miekg/dns"

	"github.com/coreos/go-systemd/daemon"
)

type Config struct {
	TLD            string
	TTL            uint32
	Listen         string
	DockerEndpoint string
	Aliases        map[string]string
}

var (
	dockerClient *docker.Client
	conf         *Config
)

func stripTLD(reqName string, tld string) string {
	name := strings.TrimSuffix(reqName, tld)
	return name[0 : len(name)-1]
}

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

func createAnswer(q dns.Question) (ans *dns.A, err error) {
	ans = new(dns.A)
	ans.Hdr = dns.RR_Header{
		Name:   q.Name,
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    conf.TTL,
	}

	ip, err := findContainerIP(stripTLD(q.Name, conf.TLD))
	if err != nil {
		return
	}
	ans.A = ip

	return
}

func handleDns(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	for _, q := range r.Question {
		ans, err := createAnswer(q)
		if err == nil {
			m.Answer = append(m.Answer, ans)
		} else {
			logrus.WithFields(logrus.Fields{"err": err, "question": q}).Error("Could not answer")
		}
	}

	if err := w.WriteMsg(m); err != nil {
		logrus.WithField("err", err).Error("Failed to write response")
	}
}

func loadConfig(path string) (conf *Config, err error) {
	conf = new(Config)

	c, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(c, conf)
	if err != nil {
		return
	}

	return
}

func main() {
	var err error
	lAddr := flag.String("listen", "127.0.0.1:8053", "Listen address")
	dAddr := flag.String("docker", "unix:///var/run/docker.sock", "Address for docker client (HTTP or Unix)")
	ttl := flag.Uint64("ttl", uint64(60), "Default TTL")
	tld := flag.String("tld", "dev.", "TLD to serve")
	loglevel := flag.String("loglevel", "info", "Logrus loglevel")
	cFile := flag.String("config", "", "Path to config file")
	systemd := flag.Bool("systemd", false, "Start as systemd service")
	createConfig := flag.Bool("createConfig", false, "Print default config file to stdout")

	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	lvl, err := logrus.ParseLevel(*loglevel)
	if err != nil {
		logrus.WithField("err", err).Fatal("Failed to parse log level")
	}
	logrus.SetLevel(lvl)

	if *cFile != "" && *createConfig == false {
		conf, err = loadConfig(*cFile)
		if err != nil {
			logrus.WithField("err", err).Fatal("Failed to load config")
		}
	} else {
		conf = new(Config)
	}

	if conf.TLD == "" {
		conf.TLD = *tld
	}

	if conf.Listen == "" {
		conf.Listen = *lAddr
	}

	if conf.DockerEndpoint == "" {
		conf.DockerEndpoint = *dAddr
	}

	if conf.TTL < uint32(*ttl) {
		conf.TTL = uint32(*ttl)
	}

	if *createConfig {
		conf.Aliases = map[string]string{"someHost.tld": "containerIdOrName"}
		bytes, _ := json.MarshalIndent(conf, "", "  ")
		fmt.Println(string(bytes))
		return
	}

	logrus.WithField("config", conf).Info("Starting server")

	dockerClient, err = docker.NewClient(conf.DockerEndpoint)
	if err != nil {
		logrus.WithField("err", err).Fatal("Failed to create docker client")
	}

	dns.HandleFunc(conf.TLD, handleDns)

	go func() {
		server := &dns.Server{Addr: conf.Listen, Net: "udp"}
		err := server.ListenAndServe()
		if err != nil {
			logrus.WithField("err", err).Fatal("Failed to setup server")
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	if *systemd == true {
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
