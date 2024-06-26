package utils

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func SplitEmailAddress(emailAddress string) ([]string, error) {
	parts := strings.Split(emailAddress, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid email address")
	}
	return parts, nil
}

// described in 4.3 MX use nslookup or dig to get the MX record
func Get_MX_record_SLD(domain string) (map[string]int, error) {

	SLD_tables := make(map[string]int)
	out, err := exec.Command("nslookup", "-type=mx ", domain).Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")

	// var mx_records []string

	re1 := regexp.MustCompile(`MX preference = (\d+)`)
	re2 := regexp.MustCompile(`mail exchanger = (.+)`)

	for _, line := range lines {
		if strings.Contains(line, "mail exchanger") {
			match1 := re1.FindStringSubmatch(line)
			match2 := re2.FindStringSubmatch(line)
			num, err := strconv.Atoi(match1[1])

			if err != nil {
				return SLD_tables, err
			}

			SLD_tables[match2[1]] = num

		}
	}

	return SLD_tables, nil
}

// func Download_publicsuffix_list(domain string) {
// 	url := "https://publicsuffix.org/list/public_suffix_list.dat"
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		os.Exit(1)
// 	}
// 	defer resp.Body.Close()

// 	// Create the file
// 	out, err := os.Create("public_suffix_list.dat")
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		os.Exit(1)
// 	}
// 	defer out.Close()

// 	// Write the body to file
// 	_, err = io.Copy(out, resp.Body)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		os.Exit(1)
// 	}
// 	return
// }

func GetPublicSuffixList() (map[string]bool, error) {
	url := "https://publicsuffix.org/list/public_suffix_list.dat"
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	tldMap := make(map[string]bool)
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 && !strings.HasPrefix(line, "//") {
			tldMap[line] = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return tldMap, nil
}

func ExtractSLD(domain string, tldMap map[string]bool) (string, error) {
	parsedURL, err := url.Parse(domain)
	if err != nil {
		return "", err
	}

	hostname := parsedURL.Hostname()
	parts := strings.Split(hostname, ".")

	if len(parts) < 2 {
		return "", fmt.Errorf("invalid domain: %s", domain)
	}

	for i := range parts {
		potentialTLD := strings.Join(parts[i:], ".")
		if tldMap[potentialTLD] {
			if i == 0 {
				return "", fmt.Errorf("no second-level domain: %s", domain)
			}
			return parts[i-1] + "." + potentialTLD, nil
		}
	}

	return "", fmt.Errorf("no public suffix found for domain: %s", domain)
}
