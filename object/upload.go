package object

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"triple-s/regex"
)

type Object struct {
	ObjectKey    string
	Size         int64
	ContentType  string
	LastModified time.Time
}

var objects []Object

func UploadObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "Invalid format. Use /{bucketName}/{objectKey}", http.StatusBadRequest)
		return
	}

	bucketName := parts[0]
	objectKey := parts[1]
	bucketPath := fmt.Sprintf("data/%s", bucketName)
	objectPath := fmt.Sprintf("%s/%s", bucketPath, objectKey)

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		http.Error(w, "Bucket does not exist", http.StatusNotFound)
		return
	}

	if !regex.IsValidBucketName(objectKey) {
		http.Error(w, "Invalid Object name", http.StatusBadRequest)
		return
	}

	file, err := os.Create(objectPath)
	if err != nil {
		http.Error(w, "Failed to save object: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	size, err := io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, "Failed to write object data", http.StatusInternalServerError)
		return
	}

	metaPath := fmt.Sprintf("%s/objects.csv", bucketPath)
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
		http.Error(w, "Failed to write metadata", http.StatusInternalServerError)
		return
	}
	defer metaFile.Close()
	for _, line := range existingLines {
		fmt.Fprintln(metaFile, line)
	}

	w.WriteHeader(http.StatusOK)
}
