package ftep

import (
	"bufio"
	"os"
	"strings"
)

// ReadAccessionsFile reads one accession per line, ignoring blank lines.
func ReadAccessionsFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var accessions []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		accessions = append(accessions, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return accessions, nil
}
