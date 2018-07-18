package server

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
