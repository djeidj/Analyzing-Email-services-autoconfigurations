package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func SplitEmailAddress(emailAddress string) []string {
	parts := strings.Split(emailAddress, "@")
	return parts
}

// described in 4.3 MX use nslookup or dig to get the MX record
func Get_MX_record_SLD(domain string) map[string]int {
	out, err := exec.Command("nslookup", "-type=mx ", domain).Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	lines := strings.Split(string(out), "\n")

	// var mx_records []string
	url_tables := make(map[string]int)

	re1 := regexp.MustCompile(`MX preference = (\d+)`)
	re2 := regexp.MustCompile(`mail exchanger = (.+)`)

	for _, line := range lines {
		if strings.Contains(line, "mail exchanger") {
			match1 := re1.FindStringSubmatch(line)
			match2 := re2.FindStringSubmatch(line)
			num, err := strconv.Atoi(match1[1])

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			url_tables[match2[1]] = num

		}
	}

	return url_tables
}
