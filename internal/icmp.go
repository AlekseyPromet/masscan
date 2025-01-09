package internal

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	protocolICMP = 1 // ICMP protocol number for IPv4
)

func (s *Server) ScanICMP(wg *sync.WaitGroup, hostname string) {

	wg.Add(1)

	go func(hostname string) {

		defer wg.Done()

		fileLogs, err := os.Create(fmt.Sprintf("%v/%v_icmp.txt",
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

		ipAddr, err := net.ResolveIPAddr("ip4", hostname)
		if err != nil {
			s.logger.Error("resolve ip", slog.String(hostname, err.Error()))
			return
		}

		s.logger.Info("resolve", slog.String("tcp", ipAddr.String()))

		conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			s.logger.Error("listen icmp", slog.String(hostname, err.Error()))
			return
		}

		_, _ = fileLogs.WriteString(hostname + " : " + time.Now().String() + "\n")
		var avg int64
		var mi int64 = 3
		for i := int64(0); i < mi; i++ {
			duration, err := s.sendIcmp(i, conn, ipAddr, fileLogs)
			if err != nil {
				continue
			}
			avg += duration

		}
		avg = avg / mi
		_, _ = fileLogs.WriteString(fmt.Sprintf("avg : %dms\n", avg))

	}(hostname)
}

func (s *Server) sendIcmp(i int64, conn *icmp.PacketConn, ipAddr *net.IPAddr, fileLogs *os.File) (int64, error) {

	// Create an ICMP Echo Request message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, // ICMP Echo Request
		Code: 0,                 // Code must be 0 for Echo Request
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,   // Use process ID as identifier
			Seq:  int(i),                 // Sequence number
			Data: []byte("Hello, ICMP!"), // Payload
		},
	}

	// Marshal the ICMP message into bytes
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		s.logger.Error("Failed to marshal ICMP message", slog.Any("err", err))
		return 0, err
	}

	s.logger.Info("Sent ICMP Echo Request to",
		slog.String(
			"ip",
			ipAddr.String(),
		))

	// Set a timeout for receiving the reply
	reply := make([]byte, 1500) // Buffer for the reply
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		s.logger.Error("Failed to set read deadline", slog.Any("err", err))
		return 0, err
	}

	// Send the ICMP Echo Request
	start := time.Now()
	if _, err := conn.WriteTo(msgBytes, ipAddr); err != nil {
		s.logger.Error("Failed to send ICMP message", slog.Any("err", err))
		return 0, err
	}

	// Receive the ICMP Echo Reply
	n, peer, err := conn.ReadFrom(reply)
	if err != nil {
		s.logger.Error("Failed to read ICMP reply", slog.Any("err", err))
		return 0, err
	}
	duration := time.Since(start)

	// Parse the ICMP reply
	replyMsg, err := icmp.ParseMessage(protocolICMP, reply[:n])
	if err != nil {
		s.logger.Error("Failed to parse ICMP reply", slog.Any("err", err))
		return 0, err
	}

	// Check if the reply is an ICMP Echo Reply
	if replyMsg.Type == ipv4.ICMPTypeEchoReply {
		echoReply, ok := replyMsg.Body.(*icmp.Echo)
		if !ok {
			s.logger.Error("Invalid ICMP Echo Reply body", slog.Any("err", err))
			return 0, err
		}
		s.logger.Info(
			"Received ICMP Echo Reply from",
			slog.Any(peer.String(), echoReply.ID),
			slog.Any(peer.String(), echoReply.Seq),
			slog.Any(peer.String(), duration),
		)

		_, err = fileLogs.WriteString(
			fmt.Sprintf("id %d \tseq %d \n%v %vms\n",
				echoReply.ID,
				echoReply.Seq,
				string(echoReply.Data),
				duration.Milliseconds()),
		)
		if err != nil {
			s.logger.Error(err.Error())
		}

		return duration.Milliseconds(), err

	}
	s.logger.Info("Received non-Echo Reply ICMP message", slog.Any("body", replyMsg.Body))

	return 0, nil
}
