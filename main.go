package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"triple-s/bucket"
	"triple-s/info"
)

type Object struct {
	ObjectKey    string
	Size         int64
	ContentType  string
	LastModified time.Time
}

var objects []Object

func main() {
	help := flag.Bool("help", false, "Prints help information")
	port := flag.Int("port", 8080, "Port number for the server")
	dir := flag.String("dir", "http://localhost", "Base URL for server")

	flag.Parse()

	if *help {
		info.PrintInfo()
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if strings.Contains(path, "/") {
			if r.Method == http.MethodPut {
				UploadObject(w, r)
				return
			}
		}
		switch r.Method {
		case http.MethodPut:
			bucket.CreateBucket(w, r)
		case http.MethodGet:
			bucket.GetAllBuckets(w, r)
		case http.MethodDelete:
			bucket.DeleteBucket(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	addr := ":" + strconv.Itoa(*port)
	fmt.Printf("Server started. Go to %s%s\n", *dir, addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}

func UploadObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/")
	paths := strings.SplitN(path, "/", 2)

	if len(paths) != 2 {
		http.Error(w, "Invalid request. Keep this format /{bucketName}/{ObjectKey}", http.StatusBadRequest)
		return
	}

	bucketName := paths[0]
	objectKey := paths[1]

	bucketPath := fmt.Sprintf("data/%s", bucketName)

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		http.Error(w, "Bucket does not exist", http.StatusNotFound)
		return
	}

	if strings.TrimSpace(objectKey) == "" || strings.Contains(objectKey, "..") {
		http.Error(w, "Invalid objectKey", http.StatusBadRequest)
		return
	}

	objects := LoadObjects(bucketName)

	if objects == nil {
		http.Error(w, "Failed to load objects from csv file", http.StatusInternalServerError)
		return
	}

	objectPath := fmt.Sprintf("%s/%s", bucketPath, objectKey)
	file, err := os.Create(objectPath)
	if err != nil {
		http.Error(w, "Failed to save object"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	size, err := io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, "Failed to copy the object content", http.StatusInternalServerError)
		return
	}

	metaPath := fmt.Sprintf("%s/object.csv", bucketPath)
	metaFile, err := os.OpenFile(metaPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		http.Error(w, "Failed to open metadata file", http.StatusInternalServerError)
		return
	}
	defer metaFile.Close()

	info, _ := metaFile.Stat()
	if info.Size() == 0 {
		metaFile.WriteString("ObjectKey,Size,ContentType,LastModified\n")
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	lastModified := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("%s,%d,%s,%s\n", objectKey, size, contentType, lastModified)

	metaFile.WriteString(line)

	w.WriteHeader(http.StatusOK)
}

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

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		http.Error(w, "Bucket does not exist", http.StatusNotFound)
		return
	}

	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		http.Error(w, "Object does not exist", http.StatusNotFound)
		return
	}

	if err := os.Remove(objectPath); err != nil {
		http.Error(w, "Failed to delete object file", http.StatusInternalServerError)
		return
	}

	err := DeleteObjectMetadata(bucketName, objectKey)
	if err != nil {
		http.Error(w, "Failed to update metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func LoadObjects(bucketName string) []Object {
	path := fmt.Sprintf("data/%s/object.csv", bucketName)
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var loaded []Object

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 4 {
			continue
		}
		size, _ := strconv.ParseInt(parts[1], 10, 64)
		t, _ := time.Parse(time.RFC3339, parts[3])
		loaded = append(loaded, Object{
			ObjectKey:    parts[0],
			Size:         size,
			ContentType:  parts[2],
			LastModified: t,
		})
	}
	return loaded
}

func DeleteObjectMetadata(bucketName, targetKey string) error {
	csvPath := fmt.Sprintf("data/%s/object.csv", bucketName)
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
		if fields[0] == targetKey {
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

	if _, err := os.Stat(bucketPath); os.IsNotExist(err) {
		http.Error(w, "Bucket does not exist", http.StatusNotFound)
		return
	}

	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}

	objects := LoadObjects(bucketName)
	contentType := "application/octet-stream"
	for _, obj := range objects {
		if obj.ObjectKey == objectKey {
			contentType = obj.ContentType
			break
		}
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
