package object

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func DeleteObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
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
		http.Error(w, "Object does not exist", http.StatusNotFound)
		return
	}

	if err := os.Remove(objectPath); err != nil {
		http.Error(w, "Failed to delete object file", http.StatusInternalServerError)
		return
	}

	metaPath := fmt.Sprintf("%s/objects.csv", bucketPath)
	lines := []string{}

	f, err := os.Open(metaPath)
	if err != nil {
		http.Error(w, "Failed to read metadata file", http.StatusInternalServerError)
		return
	}
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	lines = append(lines, scanner.Text())
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
		http.Error(w, "Failed to update metadata", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	for _, line := range lines {
		fmt.Fprintln(out, line)
	}

	w.WriteHeader(http.StatusNoContent)
}
