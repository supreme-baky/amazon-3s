package object

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func GetObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeXMLError(w, http.StatusMethodNotAllowed, "MethodNotAllowed", "Only GET method is allowed", r.URL.Path)
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

	metaPath := filepath.Join(bucketPath, "objects.csv")
	contentType := "application/octet-stream"

	if f, err := os.Open(metaPath); err == nil {
		scanner := bufio.NewScanner(f)
		scanner.Scan()
		for scanner.Scan() {
			parts := strings.Split(scanner.Text(), ",")
			if len(parts) == 4 && parts[0] == objectKey {
				contentType = parts[2]
				break
			}
		}
		f.Close()
	}

	file, err := os.Open(objectPath)
	if err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to open object file", r.URL.Path)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)
}
