package internal

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var HTTPGet = func(hostname string) []byte {
	return []byte("GET / HTTP/1.1\r\nHost: " +
		hostname + "\r\n" +
		"Connection: close\r\n" +
		"User-Agent: TLS-keys-dump" +
		"\r\n\r\n")
}

func (s *Server) ScanTCP(wg *sync.WaitGroup, hostname string) {

	wg.Add(1)

	go func(hostname string) {
		defer wg.Done()

		fileLogs, err := os.Create(fmt.Sprintf("%v/%v_tcp.txt",
			outdir,
			hostname,
		))
		if err != nil {
			s.logger.Error("file log", slog.String(hostname, err.Error()))
			return
		}
		defer func() {
			_, _ = fileLogs.WriteString("\r\n")
			fileLogs.Close()
		}()

		if s.enableTLS {
			s.scanTCPWithTLS(hostname, fileLogs)
		}
		if s.enableTCP {
			s.scanTCP(hostname, fileLogs)
		}

	}(hostname)
}

func (s *Server) scanTCPWithTLS(hostname string, fileLogs *os.File) {
	dial, err := net.DialTimeout(
		"tcp",
		hostname+":443",
		s.timeout,
	)
	if err != nil {
		s.logger.Error("err", slog.String(hostname, err.Error()))
		return
	}
	defer dial.Close()

	cfg := s.configTLS

	if s.enableLog {

		fileLogsKeys, err := createFileWithSessionTLS(hostname)
		if err != nil {
			s.logger.Error("err log file", slog.String(hostname, err.Error()))
			return
		}
		defer fileLogsKeys.Close()

		_, _ = fileLogsKeys.WriteString(time.Now().String() + "\r\n")

		cfg.KeyLogWriter = fileLogsKeys
	}

	conn := tls.Client(dial, cfg)
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(s.timeout))

	_, err = conn.Write(HTTPGet(hostname))
	if err != nil {
		s.logger.Error("conn err", slog.String(hostname, err.Error()))
		return
	}

	buffer := make([]byte, 32*1024)
	_, err = io.CopyBuffer(fileLogs, conn, buffer)
	if err != nil {
		s.logger.Error("copy err", slog.String(hostname, err.Error()))
		return
	}

	resp := make([]byte, 0, 32*1024)
	n, err := conn.Read(buffer)
	if err != nil && err != io.EOF {
		s.logger.Error("read err", slog.String(hostname, err.Error()))

		return
	}
	resp = append(resp, buffer[:n]...)

	// Check for HTTP redirect status codes (3xx)
	if bytes.HasPrefix(resp, []byte("HTTP/1.")) {
		statusLine := string(bytes.SplitN(resp, []byte("\r\n"), 2)[0])
		if strings.Contains(statusLine, " 30") {
			s.logger.Info("redirect detected",
				slog.String("host", hostname),
				slog.String("status", statusLine))
		}
	}

	s.logger.Info("tcp", slog.String(hostname, "OK"))
}

func (s *Server) scanTCP(hostname string, fileLogs *os.File) {

	_, _ = fileLogs.WriteString("=============================================\r\n")

	for serviceName, ports := range s.ports {

		for _, port := range ports {

			dial, err := net.DialTimeout(
				"tcp",
				hostname+":"+port,
				s.timeout,
			)
			if errors.Is(err, net.ErrClosed) {
				continue
			}
			if err != nil {
				s.logger.Error("err", slog.String(hostname, err.Error()))
				continue
			}
			defer dial.Close()

			// Если соединение установлено, значит порт открыт
			_, _ = fileLogs.WriteString(serviceName + " open TCP port: " + port + "\n")
			s.logger.Info("open tcp",
				slog.String("host", hostname),
				slog.String("port", port),
				slog.String("service", serviceName),
			)
		}

	}
}
