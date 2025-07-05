package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"triple-s/bucket"
	"triple-s/help/info"
	"triple-s/object"
)

func main() {
	help := flag.Bool("help", false, "Prints help information")
	port := flag.Int("port", 8080, "Port number for the server")
	dir := flag.String("dir", "http://localhost", "Base URL for server")

	flag.Parse()

	if *help {
		info.PrintInfo()
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		if strings.Contains(path, "/") {
			switch r.Method {
			case http.MethodPut:
				object.UploadObject(w, r)
			case http.MethodGet:
				object.GetObject(w, r)
			case http.MethodDelete:
				object.DeleteObject(w, r)
			default:
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
			return
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
