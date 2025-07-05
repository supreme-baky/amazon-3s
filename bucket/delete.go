package bucket

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type ErrorResponse struct {
	XMLName xml.Name `xml:"Error"`
	Code    string   `xml:"Code"`
	Message string   `xml:"Message"`
}

func DeleteBucket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeXMLError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Invalid Method. Use DELETE")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		writeXMLError(w, http.StatusBadRequest, "InvalidBucketName", "Bucket name to delete is not given.")
		return
	}

	baseDir := "data"
	bucketName := path

	bucketPath := fmt.Sprintf("%s/%s", baseDir, bucketName)

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		writeXMLError(w, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist.")
		return
	}

	metaPath := fmt.Sprintf("%s/objects.csv", bucketPath)
	if file, err := os.Open(metaPath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		lineCount := 0
		for scanner.Scan() {
			lineCount++
			if lineCount > 1 {
				writeXMLError(w, http.StatusConflict, "BucketNotEmpty", "The bucket you tried to delete is not empty.")
				return
			}
		}
	}

	if err := DeleteBucketMetadata(bucketName); err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to update bucket metadata.")
		return
	}

	if err := os.RemoveAll(bucketPath); err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to delete the bucket.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteBucketMetadata(bucketName string) error {
	csvPath := "data/buckets.csv"
	file, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, bucketName+",") {
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
		out.WriteString(line + "\n")
	}

	return nil
}

func writeXMLError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	errResp := ErrorResponse{
		Code:    code,
		Message: message,
	}
	xml.NewEncoder(w).Encode(errResp)
}
