package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// ProgressReader wraps an io.Reader and reports progress
type ProgressReader struct {
	io.Reader
	Total      int64
	Current    int64
	OnProgress func(current, total int64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)

	if pr.OnProgress != nil {
		pr.OnProgress(pr.Current, pr.Total)
	}

	return n, err
}

// DownloadWithProgress downloads a file with progress tracking
func DownloadWithProgress(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Get the content length
	size := resp.ContentLength

	// Create progress reader
	progressReader := &ProgressReader{
		Reader: resp.Body,
		Total:  size,
		OnProgress: func(current, total int64) {
			currentMB := float64(current) / 1024 / 1024
			if total > 0 {
				totalMB := float64(total) / 1024 / 1024
				percent := float64(current) / float64(total) * 100
				fmt.Printf("\rDownloading... %.2f%% (%.2f/%.2f MB)", percent, currentMB, totalMB)
			} else {
				fmt.Printf("\rDownloading... %.2f MB", currentMB)
			}
		},
	}

	// Write the body to file
	_, err = io.Copy(out, progressReader)
	if err != nil {
		return err
	}

	log.Infoln("Download complete!")
	return nil
}

// UploadWithProgress uploads data with progress tracking
func UploadWithProgress(url string, reader io.Reader, size int64) error {
	// Create progress reader
	progressReader := &ProgressReader{
		Reader: reader,
		Total:  size,
		OnProgress: func(current, total int64) {
			if total > 0 {
				percent := float64(current) / float64(total) * 100
				fmt.Printf("\rUploading... %.2f%% (%d/%d bytes)", percent, current, total)
			} else {
				fmt.Printf("\rUploading... %d bytes", current)
			}
		},
	}

	// Create the request
	req, err := http.NewRequest("POST", url, progressReader)
	if err != nil {
		return err
	}
	req.ContentLength = size

	// Send the request
	client := &http.Client{
		Timeout: 30 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	log.Infoln("Upload complete!")
	return nil
}
