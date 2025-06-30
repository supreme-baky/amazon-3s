package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Bucket struct {
	Name             string    `xml:"Name"`
	CreationTime     time.Time `xml:"CreationTime"`
	LastModifiedTime time.Time `xml:"LastModifiedTime"`
}

var buckets []Bucket

func isValidBucketName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}

	allowedChars := regexp.MustCompile(`^[a-z0-9.-]+$`)
	if !allowedChars.MatchString(name) {
		return false
	}

	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}

	if regexp.MustCompile(`[.-]{2,}`).MatchString(name) {
		return false
	}

	if ip := net.ParseIP(name); ip != nil {
		return false
	}

	return true
}

func CreateBucket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		http.Error(w, "Bucket Name is not given", http.StatusBadRequest)
		return
	}
	bucketName := path
	if isValidBucketName(bucketName) {
		http.Error(w, "Invalid Bucket name", http.StatusBadRequest)
		return
	}
	for _, bucket := range buckets {
		if bucket.Name == bucketName {
			http.Error(w, "Bucket with this name already exists", http.StatusConflict)
			return
		}
	}
	newBucket := Bucket{
		Name:             bucketName,
		CreationTime:     time.Now(),
		LastModifiedTime: time.Now(),
	}

	bucketDir := fmt.Sprintf("data/%s", bucketName)

	err := os.MkdirAll(bucketDir, 0755)
	if err != nil {
		http.Error(w, "Bucket creation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	csvPath := fmt.Sprintf("%s/buckets.csv", bucketDir)
	file, err := os.OpenFile(csvPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		http.Error(w, "Failed to write metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	info, _ := file.Stat()
	if info.Size() == 0 {
		file.WriteString("Name,CreationTime,LastModifiedTime\n")
	}

	line := fmt.Sprintf("%s,%s,%s\n",
		newBucket.Name,
		newBucket.CreationTime.Format(time.RFC3339),
		newBucket.LastModifiedTime.Format(time.RFC3339),
	)
	file.WriteString(line)

	buckets = append(buckets, newBucket)

	w.WriteHeader(http.StatusOK)
}

func GetAllBuckets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.Error(w, "Invalid URL Path", http.StatusBadRequest)
		return
	}
	type ListALLMyBuckets struct {
		XMLName xml.Name `xml:"ListAllMyBucketsResult"`
		Buckets []Bucket `xml:"Buckets>Bucket"`
	}

	response := ListALLMyBuckets{
		Buckets: buckets,
	}
	w.Header().Set("Content-Type", "application-xml")
	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(response)
}

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
	for i, b := range buckets {
		if b.Name == path {
			buckets = append(buckets[:i], buckets[i+1:]...)
			break
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	help := flag.Bool("help", false, "Prints help information")
	port := flag.Int("port", 8080, "Port number for the server")
	dir := flag.String("dir", "http://localhost", "Base URL for server")

	flag.Parse()

	if *help {
		fmt.Println(`Simple Storage Service

Usage:
  triple-s [-port <N>] [-dir <S>]
  triple-s --help

Options:
  --help       Show this help screen
  --port N     Port number to listen on (default 8080)
  --dir S      Base URL (default "http://localhost")`)
		os.Exit(0)
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
			CreateBucket(w, r)
		case http.MethodGet:
			GetAllBuckets(w, r)
		case http.MethodDelete:
			DeleteBucket(w, r)
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

	objectPath := fmt.Sprintf("%s/%s", bucketPath, objectKey)
	file, err := os.Create(objectPath)
	if err != nil {
		http.Error(w, "Failed to save object"+err.Error(), http.StatusInternalServerError)
		return
	}
}
