package bucket

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func UpdateBucketLastModified(bucketName string) error {
	csvPath := "data/buckets.csv"

	file, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	if scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	now := time.Now().Format(time.RFC3339)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) != 3 {
			lines = append(lines, line)
			continue
		}

		if parts[0] == bucketName {
			updated := fmt.Sprintf("%s,%s,%s", parts[0], parts[1], now)
			lines = append(lines, updated)
		} else {
			lines = append(lines, line)
		}
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
