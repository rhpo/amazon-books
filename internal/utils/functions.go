package utils

import (
	"fmt"
	"regexp"
)

func Report(message string, mustRun ...any) error {
	fmt.Println(message)

	if len(mustRun) > 0 && mustRun[0].(bool) == true {
		panic(message)
	}

	return fmt.Errorf("%s", message)
}

func ExtractID(url string) (string, error) {
	re := regexp.MustCompile(`/e/([A-Z0-9]+)(/|$)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", Report("ID not found in url: " + url)
}
