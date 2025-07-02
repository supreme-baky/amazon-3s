package bucket

import (
	"bufio"
	"encoding/xml"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Bucket struct {
	Name             string    `xml:"Name"`
	CreationTime     time.Time `xml:"CreationTime"`
	LastModifiedTime time.Time `xml:"LastModifiedTime"`
}

type ListAllMyBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Buckets []Bucket `xml:"Buckets>Bucket"`
}

func GetAllBuckets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.Error(w, "Invalid URL Path", http.StatusBadRequest)
		return
	}

	dataDir := "data"
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		http.Error(w, "Failed to read data directory", http.StatusInternalServerError)
		return
	}

	var allBuckets []Bucket

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		bucketName := entry.Name()
		metaPath := filepath.Join(dataDir, bucketName, "buckets.csv")

		file, err := os.Open(metaPath)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		scanner.Scan()

		if scanner.Scan() {
			parts := strings.Split(scanner.Text(), ",")
			if len(parts) != 3 {
				file.Close()
				continue
			}
			created, _ := time.Parse(time.RFC3339, parts[1])
			modified, _ := time.Parse(time.RFC3339, parts[2])
			allBuckets = append(allBuckets, Bucket{
				Name:             parts[0],
				CreationTime:     created,
				LastModifiedTime: modified,
			})
		}
		file.Close()
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(ListAllMyBucketsResult{
		Buckets: allBuckets,
	})
}
