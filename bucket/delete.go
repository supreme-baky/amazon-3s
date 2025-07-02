package bucket

import (
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		writeXMLError(w, http.StatusBadRequest, "InvalidBucketName", "Bucket name to delete is not given")
		return
	}

	bucketPath := fmt.Sprintf("data/%s", path)

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		writeXMLError(w, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist.")
		return
	}

	objectMeta := fmt.Sprintf("%s/objects.csv", bucketPath)
	if info, err := os.Stat(objectMeta); err == nil && info.Size() > 0 {
		f, _ := os.Open(objectMeta)
		defer f.Close()
		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		if n > 0 && strings.Count(string(buf), "\n") > 1 {
			writeXMLError(w, http.StatusConflict, "BucketNotEmpty", "The bucket you tried to delete is not empty.")
			return
		}
	}

	if err := os.RemoveAll(bucketPath); err != nil {
		writeXMLError(w, http.StatusInternalServerError, "InternalError", "Failed to delete the bucket.")
		return
	}

	DeleteBucketMetadata(path)

	w.WriteHeader(http.StatusNoContent)
}

func DeleteBucketMetadata(bucketName string) error {
	csvPath := fmt.Sprintf("data/%s/buckets.csv", bucketName)
	if _, err := os.Stat(csvPath); err == nil {
		return os.Remove(csvPath)
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
