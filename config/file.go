package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var emptyLine = regexp.MustCompile(`#(.*)$`)

func (config *Config) parseFile(filename string) error {

	fd, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		lineWithoutComments := strings.TrimSpace(emptyLine.ReplaceAllString(line, ""))
		if lineWithoutComments == `` || lineWithoutComments == "#" {
			continue
		}

		key, value := ``, ``
		parsed := false

		// если key == value
		if strings.Index(lineWithoutComments, `=`) != -1 {
			configLine := strings.Split(lineWithoutComments, `=`)
			if len(configLine) == 1 || len(configLine) > 2 {
				return fmt.Errorf("Set config like 'key = value' in line: `%s` \n", line)
			}
			key, value = configLine[0], configLine[1]
			parsed = true
		}

		// если key value
		if !parsed && strings.Index(lineWithoutComments, ` `) != -1 {
			configLine := strings.Split(lineWithoutComments, " ")
			if len(configLine) == 1 || len(configLine) > 2 {
				return fmt.Errorf("Set config like 'key value' in line: `%s` \n", line)
			}
			key = configLine[0]
			valueArr := configLine[1:len(configLine)]
			value = strings.Join(valueArr, " ")

		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if err := config.set(key, value); err != nil {
			return err
		}
	}

	return nil
}
