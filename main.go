package main

import "github.com/AlekseyPromet/masscan/internal"

var (
	Version   string
)

func main() {

	println(Version)

	opts := []internal.ServerOpt{}
	opts = append(opts,
		// TODO: add uri address for scan services
		internal.ServerSetTargets(nil), // []string{"google.com", "aamazon.com", etc}
		internal.ServerLogEnabled,
		internal.ServerTLSEnabled,
		internal.ServerUDPEnabled,
		internal.ServerSetTimeout,
		internal.ServerSetPorts,
		internal.ServerTCPEnabled,
	)

	srv := internal.NewServer(opts...)
	if err := srv.Scann(); err != nil {
		panic(err)
	}
}
