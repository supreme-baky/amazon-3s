package bucket

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"triple-s/help/regex"
)

type BucketInfo struct {
	XMLName      xml.Name `xml:"Bucket"`
	Name         string   `xml:"Name"`
	CreationTime string   `xml:"CreationTime"`
	LastModified string   `xml:"LastModifiedTime"`
}

type XMLError struct {
	XMLName xml.Name `xml:"Error"`
	Code    string   `xml:"Code"`
	Message string   `xml:"Message"`
}

func CreateBucket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeXMLError(w, http.StatusMethodNotAllowed, "MethodNotAllowed", "Only PUT method is allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		writeXMLError(w, http.StatusBadRequest, "MissingBucketName", "Bucket name is required")
		return
	}

	bucketName := path
	baseDir := "data"

	if !regex.IsValidBucketName(bucketName) {
		writeXMLError(w, http.StatusBadRequest, "InvalidBucketName", "Bucket name is invalid")
		return
	}

	bucketDir := fmt.Sprintf("%s/%s", baseDir, bucketName)
	if _, err := os.Stat(bucketDir); err == nil {
		writeXMLError(w, http.StatusConflict, "BucketAlreadyExists", "A bucket with this name already exists")
		return
	}

	if err := os.MkdirAll(bucketDir, 0o755); err != nil {
		writeXMLError(w, http.StatusInternalServerError, "BucketCreationFailed", err.Error())
		return
	}

	csvPath := fmt.Sprintf("%s/buckets.csv", baseDir)
	header := ""
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		header = "Name,CreationTime,LastModifiedTime\n"
	}

	now := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("%s,%s,%s\n", bucketName, now, now)

	file, err := os.OpenFile(csvPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		writeXMLError(w, http.StatusInternalServerError, "MetadataWriteFailed", "Failed to open buckets.csv: "+err.Error())
		return
	}
	defer file.Close()

	if header != "" {
		if _, err := file.WriteString(header); err != nil {
			writeXMLError(w, http.StatusInternalServerError, "MetadataHeaderWriteFailed", err.Error())
			return
		}
	}

	if _, err := file.WriteString(line); err != nil {
		writeXMLError(w, http.StatusInternalServerError, "MetadataAppendFailed", err.Error())
		return
	}

	response := BucketInfo{
		Name:         bucketName,
		CreationTime: now,
		LastModified: now,
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(response)
}
