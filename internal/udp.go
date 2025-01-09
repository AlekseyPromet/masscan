package internal

import (
	"log/slog"
	"net"
	"os"
	"sync"
	"time"
)

func (s *Server) UDPScan(wg *sync.WaitGroup, target string) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		fileLogs, err := os.Create(outdir + "/" + target + "_udp.txt")
		if err != nil {
			s.logger.Error("file log", slog.String(target, err.Error()))
			return
		}
		defer func() {
			_, _ = fileLogs.WriteString("\r\n")
			fileLogs.Close()
		}()

		for serviceName, ports := range s.ports {

			for _, port := range ports {
				s.logger.Info("scan udp",
					slog.String("host", target),
					slog.String("port", port),
				)
				addr := target + ":" + port
				conn, err := net.DialTimeout("udp", addr, s.timeout)
				if err != nil {
					s.logger.Error("connect udp",
						slog.String("host", target),
						slog.String("port", port),
						slog.String("error", err.Error()),
					)
					continue
				}
				defer conn.Close()

				_, _ = fileLogs.WriteString(serviceName + " open UDP port: " + port + "\n")
				s.logger.Info("open udp",
					slog.String("host", target),
					slog.String("port", port),
					slog.String("service", serviceName),
				)

				// Отправляем тестовый пакет
				_, err = conn.Write([]byte("UDP Test\n"))
				if err != nil {
					s.logger.Error("write udp",
						slog.String("host", target),
						slog.String("port", port),
						slog.String("error", err.Error()),
					)
					continue
				}

				// Устанавливаем таймаут для чтения ответа
				_ = conn.SetReadDeadline(time.Now().Add(s.timeout))

				buffer := make([]byte, 1024)
				n, err := conn.Read(buffer)
				if err != nil {
					if err, ok := err.(net.Error); ok && err.Timeout() {
						// Таймаут - порт может быть открыт, но не отвечает
						continue
					}
					s.logger.Error("read udp",
						slog.String("host", target),
						slog.String("port", port),
						slog.String("error", err.Error()),
					)
					continue
				}

				_, _ = fileLogs.WriteString("Response from port " + port + ": " + string(buffer[:n]) + "\n")
			}
		}

	}()

}
