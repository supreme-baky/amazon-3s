package bucket

import (
	"bufio"
	"encoding/xml"
	"net/http"
	"os"
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
		writeXMLError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Invalid Method. Use GET")
		return
	}

	if r.URL.Path != "/" {
		writeXMLError(w, http.StatusBadRequest, "Invalid URL Path", "Type valid URL")
		return
	}

	file, err := os.Open("data/buckets.csv")
	if err != nil {
		writeXMLError(w, http.StatusInternalServerError, "Failed to open buckets.csv", "Check bucket.csv inside ./data folder")
		return
	}
	defer file.Close()

	var allBuckets []Bucket
	scanner := bufio.NewScanner(file)

	if scanner.Scan() {
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.Split(line, ",")
			if len(parts) != 3 {
				continue
			}
			created, err1 := time.Parse(time.RFC3339, parts[1])
			modified, err2 := time.Parse(time.RFC3339, parts[2])
			if err1 != nil || err2 != nil {
				continue
			}
			allBuckets = append(allBuckets, Bucket{
				Name:             parts[0],
				CreationTime:     created,
				LastModifiedTime: modified,
			})
		}
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(ListAllMyBucketsResult{
		Buckets: allBuckets,
	})
}
