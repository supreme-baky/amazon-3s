package bucket

import (
	"encoding/xml"
	"net/http"
)

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
