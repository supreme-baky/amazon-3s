package object

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"triple-s/bucket"
	"triple-s/help/regex"
)

type Object struct {
	ObjectKey    string
	Size         int64
	ContentType  string
	LastModified string
}

var objects []Object

type XMLError struct {
	XMLName   string `xml:"Error"`
	Code      string `xml:"Code"`
	Message   string `xml:"Message"`
	Resource  string `xml:"Resource"`
	RequestID string `xml:"RequestId"`
}

func writeXMLError(w http.ResponseWriter, statusCode int, code, message, resource string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)

	errResp := XMLError{
		Code:      code,
		Message:   message,
		Resource:  resource,
		RequestID: "1234567890",
	}
	xml.NewEncoder(w).Encode(errResp)
}

func UploadObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeXMLError(w, http.StatusMethodNotAllowed, "MethodNotAllowed", "Only PUT is allowed", r.URL.Path)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		writeXMLError(w, http.StatusBadRequest, "InvalidURI", "Expected format: /{bucketName}/{objectKey}", r.URL.Path)
		return
	}

	bucketName := parts[0]
	objectKey := parts[1]
	baseDir := "data"
	bucketPath := filepath.Join(baseDir, bucketName)
	objectPath := filepath.Join(bucketPath, objectKey)

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		writeXMLError(w, http.StatusNotFound, "NoSuchBucket", "Bucket does not exist", r.URL.Path)
		return
	}

	if !regex.IsValidBucketName(objectKey) {
		writeXMLError(w, http.StatusBadRequest, "InvalidObjectName", "Invalid object name", r.URL.Path)
		return
	}

	file, err := os.Create(objectPath)
	if err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to save object", r.URL.Path)
		return
	}
	defer file.Close()

	size, err := io.Copy(file, r.Body)
	if err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to write object data", r.URL.Path)
		return
	}

	metaPath := filepath.Join(bucketPath, "objects.csv")
	existingLines := []string{"ObjectKey,Size,ContentType,LastModified"}

	if f, err := os.Open(metaPath); err == nil {
		scanner := bufio.NewScanner(f)
		scanner.Scan()
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.SplitN(line, ",", 2)
			if fields[0] != objectKey {
				existingLines = append(existingLines, line)
			}
		}
		f.Close()
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	lastModified := time.Now().Format(time.RFC3339)
	newLine := fmt.Sprintf("%s,%d,%s,%s", objectKey, size, contentType, lastModified)
	existingLines = append(existingLines, newLine)

	metaFile, err := os.Create(metaPath)
	if err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to write metadata", r.URL.Path)
		return
	}
	defer metaFile.Close()

	for _, line := range existingLines {
		fmt.Fprintln(metaFile, line)
	}

	if err := bucket.UpdateBucketLastModified(bucketName); err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to update bucket metadata", r.URL.Path)
		return
	}

	response := Object{
		ObjectKey:    objectKey,
		Size:         size,
		ContentType:  contentType,
		LastModified: lastModified,
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(response)
}
