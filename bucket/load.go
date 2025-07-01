package bucket

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func LoadBuckets(bucketName string) []Bucket {
	path := fmt.Sprintf("data/%s/buckets.csv", bucketName)
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var loaded []Bucket

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 3 {
			continue
		}
		creationT, _ := time.Parse(time.RFC3339, parts[1])
		lModifiedT, _ := time.Parse(time.RFC3339, parts[2])

		loaded = append(loaded, Bucket{
			Name:             parts[0],
			CreationTime:     creationT,
			LastModifiedTime: lModifiedT,
		})
	}
	return loaded
}
