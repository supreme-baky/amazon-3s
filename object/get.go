package object

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}

	metaPath := fmt.Sprintf("%s/objects.csv", bucketPath)
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
		http.Error(w, "Error opening object file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)
}
