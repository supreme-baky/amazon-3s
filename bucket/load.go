package bucket

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func LoadBuckets(bucketName string) []Bucket {
	csvPath := fmt.Sprintf("data/%s/buckets.csv", bucketName)
	file, err := os.Open(csvPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	if !scanner.Scan() {
		return nil
	}

	parts := strings.Split(scanner.Text(), ",")
	if len(parts) != 3 {
		return nil
	}

	created, _ := time.Parse(time.RFC3339, parts[1])
	modified, _ := time.Parse(time.RFC3339, parts[2])

	return []Bucket{{
		Name:             parts[0],
		CreationTime:     created,
		LastModifiedTime: modified,
	}}
}
