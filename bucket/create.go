package bucket

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"triple-s/regex"
)

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

	if !regex.IsValidBucketName(bucketName) {
		http.Error(w, "Invalid Bucket name", http.StatusBadRequest)
		return
	}

	bucketDir := fmt.Sprintf("data/%s", bucketName)
	if _, err := os.Stat(bucketDir); err == nil {
		http.Error(w, "Bucket already exists", http.StatusConflict)
		return
	}

	if err := os.MkdirAll(bucketDir, 0755); err != nil {
		http.Error(w, "Failed to create bucket directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	meta := fmt.Sprintf("Name,CreationTime,LastModifiedTime\n%s,%s,%s\n",
		bucketName, now.Format(time.RFC3339), now.Format(time.RFC3339))

	csvPath := fmt.Sprintf("%s/buckets.csv", bucketDir)
	if err := os.WriteFile(csvPath, []byte(meta), 0644); err != nil {
		http.Error(w, "Failed to write metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
