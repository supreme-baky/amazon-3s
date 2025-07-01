package bucket

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func DeleteBucket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		http.Error(w, "Bucket Name to delete is not given", http.StatusBadRequest)
		return
	}

	bucketPath := fmt.Sprintf("data/%s", path)

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		http.Error(w, "Bucket does not exist", http.StatusNotFound)
		return
	}
	files, err := os.ReadDir(bucketPath)
	if err != nil {
		http.Error(w, "Failed to read bucket contents", http.StatusInternalServerError)
		return
	}

	for _, f := range files {
		if f.Name() == "object.csv" {
			http.Error(w, "Bucket must be empty", http.StatusConflict)
			return
		}
	}

	if err := os.RemoveAll(bucketPath); err != nil {
		http.Error(w, "Failed to delete files", http.StatusInternalServerError)
		return
	}

	buckets := LoadBuckets(path)
	for _, b := range buckets {
		if b.Name == path {
			err := DeleteBucketMetadata(path)
			if err != nil {
				http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteBucketMetadata(bucketName string) error {
	csvPath := fmt.Sprintf("data/%s/buckets.csv", bucketName)
	file, err := os.Open(csvPath)
	if err != nil {
		return err
	}

	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)

	isFirstLine := true
	for scanner.Scan() {
		line := scanner.Text()
		if isFirstLine {
			lines = append(lines, line)
			isFirstLine = false
			continue
		}
		fields := strings.SplitN(line, ",", 2)
		if fields[0] == bucketName {
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	out, err := os.Create(csvPath)
	if err != nil {
		return err
	}
	defer out.Close()

	for _, line := range lines {
		_, err := out.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}
