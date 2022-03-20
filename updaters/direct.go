package updaters

import (
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

type Direct struct {
	url  string
	seed string
}

func NewDirectUpdater(url string, seed string) (*Direct, error) {
	return &Direct{
		url:  url,
		seed: seed,
	}, nil
}

func (d *Direct) Method() string {
	return "direct"
}

func (d *Direct) Lookup() (updateAvailable bool, err error) {
	outputFile := d.getOutputFileName()
	if d.seed == outputFile {
		return false, nil
	} else {
		return true, nil
	}
}

func (d *Direct) Download() (output string, err error) {
	output = d.getOutputFileName()
	err = downloadFile(output, d.url)

	return
}

func (d *Direct) getOutputFileName() string {
	urlLastPartStart := strings.LastIndex(d.url, "/")
	if urlLastPartStart == -1 {
		urlLastPartStart = 0
	}

	urlArgumentsStart := strings.LastIndex(d.url, "?")
	if urlArgumentsStart == -1 {
		urlArgumentsStart = len(d.url)
	}

	fileName := d.url[urlLastPartStart:urlArgumentsStart]

	return filepath.Dir(d.seed) + "/" + fileName
}

func downloadFile(filePath string, url string) (err error) {
	output, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer output.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading",
	)

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan

		_ = resp.Body.Close()
		_ = output.Close()
		_ = os.Remove(filePath)

		os.Exit(0)
	}()

	_, err = io.Copy(io.MultiWriter(output, bar), resp.Body)
	return err
}
