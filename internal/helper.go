package internal

import (
	"errors"
	"log/slog"
	"os"
)

const outdir = "scan_out"

func init() {
	if err := os.Mkdir(outdir, 0666); errors.Is(err, os.ErrExist) {
		slog.Info("dir is exist")
	} else if err != nil {
		panic(err)
	}
}

func createFileWithSessionTLS(target string) (*os.File, error) {
	return os.Create(outdir + "/" + target + "_keys.txt")
}
