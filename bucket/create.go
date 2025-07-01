package bucket

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"triple-s/regex"
)

type Bucket struct {
	Name             string    `xml:"Name"`
	CreationTime     time.Time `xml:"CreationTime"`
	LastModifiedTime time.Time `xml:"LastModifiedTime"`
}

var buckets []Bucket

func CreateBucket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		http.Error(w, "Bucket Name is not given", http.StatusBadRequest)
		return
	}
	bucketName := path
	if regex.IsValidBucketName(bucketName) {
		http.Error(w, "Invalid Bucket name", http.StatusBadRequest)
		return
	}
	for _, bucket := range buckets {
		if bucket.Name == bucketName {
			http.Error(w, "Bucket with this name already exists", http.StatusConflict)
			return
		}
	}
	newBucket := Bucket{
		Name:             bucketName,
		CreationTime:     time.Now(),
		LastModifiedTime: time.Now(),
	}

	bucketDir := fmt.Sprintf("data/%s", bucketName)

	err := os.MkdirAll(bucketDir, 0755)
	if err != nil {
		http.Error(w, "Bucket creation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	csvPath := fmt.Sprintf("%s/buckets.csv", bucketDir)
	file, err := os.OpenFile(csvPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		http.Error(w, "Failed to write metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	info, _ := file.Stat()
	if info.Size() == 0 {
		file.WriteString("Name,CreationTime,LastModifiedTime\n")
	}

	line := fmt.Sprintf("%s,%s,%s\n",
		newBucket.Name,
		newBucket.CreationTime.Format(time.RFC3339),
		newBucket.LastModifiedTime.Format(time.RFC3339),
	)
	file.WriteString(line)

	w.WriteHeader(http.StatusOK)
}
