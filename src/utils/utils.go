package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func Split_EmailAddress(emailAddress string) ([]string, error) {
	parts := strings.Split(emailAddress, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid email address")
	}
	return parts, nil
}

// described in 4.3 MX use nslookup or dig to get the MX record
// `Get_MX_record_SLD` call `ExtractSLD_localpuffixlist` and `loadMapFromFile` and `SortMapByValue`
func Get_MX_full_main_domain(domain string, suffixlistpath string) ([2]string, error) {
	tldMap, err := loadMapFromFile(suffixlistpath)
	if err != nil {
		return [2]string{"", ""}, err
	}

	MX_full_main_tables := make(map[[2]string]int)
	out, err := exec.Command("nslookup", "-type=mx ", domain).Output()
	if err != nil {
		return [2]string{"", ""}, err
	}

	// every line is ended with "\r\n"
	lines := strings.Split(string(out), "\r\n")

	// var mx_records []string

	re1 := regexp.MustCompile(`MX preference = (\d+)`)
	re2 := regexp.MustCompile(`mail exchanger = (.+)`)

	for _, line := range lines {
		if strings.Contains(line, "mail exchanger") {
			match1 := re1.FindStringSubmatch(line)
			match2 := re2.FindStringSubmatch(line)

			num, err := strconv.Atoi(match1[1])
			if err != nil {
				return [2]string{"", ""}, err
			}

			mxmaindomain, err := Extract_SLDFromTLDmap(match2[1], tldMap)
			if err != nil {
				return [2]string{"", ""}, err
			}

			tmp1 := strings.Split(match2[1], ".")
			mxfulldomain := strings.Join(tmp1[1:], ".")

			MX_full_main_tables[[2]string{mxfulldomain, mxmaindomain}] = num
		}
	}

	// sort the map by value
	MX_full_main_slice := SortMapByValue(MX_full_main_tables)

	// Only the highest priority MX hostname is considered
	return MX_full_main_slice[0], nil
}

// Extract the second-level domain from a domain name using tldmap
func Extract_SLDFromTLDmap(domain string, tldMap map[string]bool) (string, error) {
	parts := strings.Split(domain, ".")

	if len(parts) < 2 {
		return "", fmt.Errorf("invalid domain: %s", domain)
	}

	for i := range parts[0:] {
		potentialTLD := strings.Join(parts[i:], ".")
		fmt.Println(potentialTLD)
		if tldMap[potentialTLD] {
			if i == 0 {
				return "", fmt.Errorf("no second-level domain: %s", domain)
			}
			return parts[i-1] + "." + potentialTLD, nil
		}
	}

	return "", fmt.Errorf("no public suffix found for domain: %s", domain)
}

// load a file in format of json to a map[string]bool
func loadMapFromFile(suffixlistpath string) (map[string]bool, error) {
	file, err := os.Open(suffixlistpath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var m map[string]bool
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&m)
	return m, err
}

// sort a map[[2]string]int by value
func SortMapByValue(m map[[2]string]int) [][2]string {
	var keys [][2]string
	for key := range m {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] < m[keys[j]]
	})
	return keys
}

// download the autoconfig.xml file to xmlpath
func Get_AutoconfigXML(url string, xmlpath string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		outFile, err := os.OpenFile(xmlpath, os.O_RDWR, 0755)
		if err != nil {
			return fmt.Errorf("error opening file: %v", xmlpath)
		}

		defer outFile.Close()

		_, err = io.Copy(outFile, response.Body)
		if err != nil {
			return fmt.Errorf("error saving to file: %v", xmlpath)
		}

		return nil
	}

	return fmt.Errorf("error downloading file: %v", url)

}
