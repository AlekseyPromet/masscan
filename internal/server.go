package internal

import (
	"crypto/tls"
	"log/slog"
	"sync"
	"time"
)

type Server struct {
	enableTCP  bool
	enableICMP bool
	enableTLS  bool
	enableLog  bool
	enableUDP  bool
	targets    []string
	ports      map[string][]string
	configTLS  *tls.Config
	timeout    time.Duration
	logger     slog.Logger
}

type ServerOpt func(s *Server)

var ServerTLSEnabled = func(s *Server) {
	s.enableTLS = true
	s.enableTCP = true
}

var ServerLogEnabled = func(s *Server) {
	s.enableLog = true
}

var ServerSetTimeout = func(s *Server) {
	s.timeout = time.Duration(time.Second * 3)
}

var ServerSetPorts = func(s *Server) {
	s.ports = make(map[string][]string, 0)

	s.ports["ssh"] = []string{"22"}
	s.ports["redis"] = []string{"6379"}
	s.ports["PostgreSQL"] = []string{"5432"}
	s.ports["MySQL"] = []string{"3306"}
	s.ports["Memcached"] = []string{"11211"}
	s.ports["ClickHouse"] = []string{"8123", "9000"}
	s.ports["HTTP"] = []string{"80"}
	s.ports["HTTPS"] = []string{"443"}
	s.ports["MongoDB"] = []string{"27017"}
	s.ports["Elasticsearch"] = []string{"9200"}
	s.ports["RabbitMQ"] = []string{"5672"}
	s.ports["Kafka"] = []string{"9092"}
	s.ports["DNS"] = []string{"53"}
	s.ports["SMTP"] = []string{"25", "587"}
	s.ports["IMAP"] = []string{"143", "993"}
	s.ports["POP3"] = []string{"110", "995"}
	s.ports["FTP"] = []string{"20", "21"}
	s.ports["LDAP"] = []string{"389", "636"}
	s.ports["NTP"] = []string{"123"}
	s.ports["SNMP"] = []string{"161", "162"}
	s.ports["Cassandra"] = []string{"9042"}
	s.ports["Neo4j"] = []string{"7687"}
	s.ports["Consul"] = []string{"8500"}
	s.ports["etcd"] = []string{"2379", "2380"}
	s.ports["Prometheus"] = []string{"9090"}
	s.ports["Grafana"] = []string{"3000"}
	s.ports["Jenkins"] = []string{"8080"}
	s.ports["Kubernetes-API"] = []string{"6443"}
	s.ports["Kubernetes-Dashboard"] = []string{"8001"}
	s.ports["Zookeeper"] = []string{"2181"}
	s.ports["RDP"] = []string{"3389"}
	s.ports["VNC"] = []string{"5900"}
	s.ports["Solr"] = []string{"8983"}
	s.ports["Gitlab"] = []string{"80", "443", "22"}
	s.ports["Docker"] = []string{"2375", "2376"}
	s.ports["Nomad"] = []string{"4646"}
	s.ports["Vault"] = []string{"8200"}
	s.ports["MinIO"] = []string{"9000", "9001"}
	s.ports["Nexus"] = []string{"8081"}
	s.ports["SonarQube"] = []string{"9000"}
	s.ports["Kibana"] = []string{"5601"}
	s.ports["Logstash"] = []string{"5044", "9600"}
	s.ports["Traefik"] = []string{"8080", "80"}
	s.ports["HAProxy"] = []string{"80", "443", "1024"}
}

var ServerSetTargets = func(targets []string) func(s *Server) {
	return func(s *Server) {
		if len(targets) == 0 {
			s.targets = []string{"ya.ru", "m.vk.com", "google.com"}
			return
		}

		s.targets = make([]string, 0, len(targets))
		s.targets = append(s.targets, targets...)
	}
}

var ServerUDPEnabled = func(s *Server) {
	s.enableUDP = true
}

var ServerTCPEnabled = func(s *Server) {
	s.enableTCP = true
}

func NewServer(opts ...ServerOpt) *Server {
	s := &Server{
		enableTLS: true,
		logger:    *slog.Default(),
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.enableTLS {
		cfg := &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS13,
			MaxVersion:         tls.VersionTLS13,
			RootCAs:            nil,
		}

		s.configTLS = cfg
	}

	return s
}

func (s *Server) Scann() error {

	wg := &sync.WaitGroup{}

	for _, target := range s.targets {

		if s.enableTCP {
			s.ScanTCP(wg, target)
		}

		if s.enableICMP {
			s.ScanICMP(wg, target)
		}

		if s.enableUDP {
			s.UDPScan(wg, target)
		}

	}

	wg.Wait()

	return nil
}
