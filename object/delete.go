package object

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"triple-s/bucket"
)

func DeleteObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeXMLError(w, http.StatusMethodNotAllowed, "MethodNotAllowed", "Only DELETE method is allowed", r.URL.Path)
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

	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		writeXMLError(w, http.StatusNotFound, "NoSuchKey", "The specified object does not exist", r.URL.Path)
		return
	}

	if err := os.Remove(objectPath); err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to delete object file", r.URL.Path)
		return
	}

	metaPath := filepath.Join(bucketPath, "objects.csv")
	lines := []string{}

	f, err := os.Open(metaPath)
	if err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to read metadata file", r.URL.Path)
		return
	}
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.SplitN(line, ",", 2)
		if fields[0] != objectKey {
			lines = append(lines, line)
		}
	}
	f.Close()

	out, err := os.Create(metaPath)
	if err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to update metadata", r.URL.Path)
		return
	}
	defer out.Close()
	for _, line := range lines {
		fmt.Fprintln(out, line)
	}

	err = bucket.UpdateBucketLastModified(bucketName)
	if err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to update bucket metadata", r.URL.Path)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
