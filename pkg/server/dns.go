package server

import (
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"strings"
	"github.com/overlordtm/docker-dns/pkg/docker"
	"github.com/overlordtm/docker-dns/pkg/config"
)

type Server struct {
	finder    docker.DockerFinder
	dnsServer *dns.Server
	conf      *config.Config
}

func New(conf *config.Config) *Server {

	finder, err := docker.New(conf)

	if err != nil {
		logrus.WithError(err).Fatal("Fatal error")
	}

	dnsServer := &dns.Server{Addr: conf.Listen, Net: "udp"}

	s := &Server{
		finder:    finder,
		dnsServer: dnsServer,
		conf:      conf,
	}

	dns.HandleFunc(conf.TLD, s.handleDns)

	return s
}

func (s *Server) Handle() {
	go func() {
		server := &dns.Server{Addr: s.conf.Listen, Net: "udp"}
		err := server.ListenAndServe()
		if err != nil {
			logrus.WithField("err", err).Fatal("Failed to setup server")
		}
	}()
}

func (s *Server) createAnswer(q dns.Question) (ans *dns.A, err error) {
	ans = new(dns.A)
	ans.Hdr = dns.RR_Header{
		Name:   q.Name,
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    s.conf.TTL,
	}

	ip, err := s.finder.FindContainerIP(stripTLD(q.Name, s.conf.TLD))
	if err != nil {
		return
	}
	ans.A = ip

	return
}

func (s *Server) handleDns(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	for _, q := range r.Question {
		ans, err := s.createAnswer(q)
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

func stripTLD(reqName string, tld string) string {
	name := strings.TrimSuffix(reqName, tld)
	return name[0 : len(name)-1]
}
