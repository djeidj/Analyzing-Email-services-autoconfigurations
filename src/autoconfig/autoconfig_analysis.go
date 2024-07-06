package autoconfig

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Download_AutoconfigXML(email_address string, suffixlistpath string, path string) error {
	// download the url_list's XML file to path
	xmlpath := filepath.Join(path, email_address+".xml")

	dir := filepath.Dir(xmlpath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", dir)
	}

	parts := strings.Split(email_address, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email address: %v", email_address)
	}

	email_domain := parts[1]
	//email_local := parts[0]

	url_list := make([]string, 0)

	// 1.1. https://autoconfig.%EMAILDOMAIN%/mail/config-v1.1.xml?emailaddress=%EMAILADDRESS% (Required)
	url_1_1 := "https://autoconfig." + email_domain + "/mail/config-v1.1.xml?emailaddress=" + email_address
	url_list = append(url_list, url_1_1)

	// 1.2. https://%EMAILDOMAIN%/.well-known/autoconfig/mail/config-v1.1.xml (Recommended)
	url_1_2 := "https://" + email_domain + "/.well-known/autoconfig/mail/config-v1.1.xml"
	url_list = append(url_list, url_1_2)

	// 1.3. http://autoconfig.%EMAILDOMAIN%/mail/config-v1.1.xml(Optional)
	url_1_3 := "http://autoconfig." + email_domain + "/mail/config-v1.1.xml"
	url_list = append(url_list, url_1_3)

	// 2.1. %ISPDB%%EMAILDOMAIN% (Recommended)
	// %ISPDB% = https://autoconfig.thunderbird.net/v1.1/
	url_2_1 := "https://autoconfig.thunderbird.net/v1.1/" + email_domain
	url_list = append(url_list, url_2_1)

	// 3
	// 1. you need to download the sufficlist first use `Get_PublicSuffixList`
	// 2. use `Get_MX_full_main_domain` to get mxfulldomain and mxmaindomain

	mx_full_main_domain, err := Get_MX_full_main_domain(email_domain, suffixlistpath)

	// if there is no MX record, dont return, continue
	if err == nil {
		mxfulldomain := mx_full_main_domain[0]
		mxmaindomain := mx_full_main_domain[1]

		// 3.1 https://autoconfig.%MXFULLDOMAIN%/mail/config-v1.1.xml?emailaddress=%EMAILADDRESS% (Recommended)
		url_3_1 := "https://autoconfig." + mxfulldomain + "/mail/config-v1.1.xml?emailaddress=" + email_address
		url_list = append(url_list, url_3_1)

		// 3.2 https://autoconfig.%MXMAINDOMAIN%/mail/config-v1.1.xml?emailaddress=%EMAILADDRESS% (Recommended)
		url_3_2 := "https://autoconfig." + mxmaindomain + "/mail/config-v1.1.xml?emailaddress=" + email_address
		url_list = append(url_list, url_3_2)

		// 3.3 %ISPDB%%MXFULLDOMAIN% (Recommended)
		url_3_3 := "https://autoconfig.thunderbird.net/v1.1/" + mxfulldomain
		url_list = append(url_list, url_3_3)

		// 3.4 %ISPDB%%MXMAINDOMAIN% (Recommended)
		url_3_4 := "https://autoconfig.thunderbird.net/v1.1/" + mxmaindomain
		url_list = append(url_list, url_3_4)
	}

	for _, url := range url_list {
		err := Get_AutoconfigXML(url, xmlpath)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("can't find Autoconfigxml file for %v", email_address)

}

// download the autoconfig.xml (use GET) file to xmlpath
func Get_AutoconfigXML(url string, xmlpath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		outFile, err := os.Create(xmlpath)
		if err != nil {
			return fmt.Errorf("error creating file: %v", err)
		}

		defer outFile.Close()

		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			return fmt.Errorf("error saving to file: %v", xmlpath)
		}

		return nil
	}

	return fmt.Errorf("error downloading file: %v", url)

}

// Save the public suffix list to a file in fomat of json
func Get_PublicSuffixList(suffixlistpath string) error {
	url := "https://publicsuffix.org/list/public_suffix_list.dat"
	response, err := http.Get(url)
	if err != nil {
		return err
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
		return err
	}

	dir := filepath.Dir(suffixlistpath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", dir)
	}

	file, err := os.Create(suffixlistpath)
	if err != nil {
		return fmt.Errorf("failed to create file: %s", suffixlistpath)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(tldMap); err != nil {
		return fmt.Errorf("failed to write to file: %s", suffixlistpath)
	}
	return nil
}

// `Get_MX_record_SLD` call `ExtractSLD_localpuffixlist` and `loadMapFromFile`
func Get_MX_full_main_domain(domain string, suffixlistpath string) ([2]string, error) {
	tldMap, err := loadMapFromFile(suffixlistpath)
	if err != nil {
		return [2]string{"", ""}, err
	}

	mx, err := net.LookupMX(domain)
	if err != nil {
		return [2]string{"", ""}, err
	}

	mxmaindomian, err := Extract_SLDFromTLDmap(mx[0].Host, tldMap)
	if err != nil {
		return [2]string{"", ""}, err

	}

	tmp1 := strings.Split(mx[0].Host, ".")
	mxfulldomain := strings.Join(tmp1[1:], ".")

	return [2]string{mxfulldomain, mxmaindomian}, nil
}

// Extract the second-level domain from a domain name using tldmap
func Extract_SLDFromTLDmap(domain string, tldMap map[string]bool) (string, error) {
	parts := strings.Split(domain, ".")

	if len(parts) < 2 {
		return "", fmt.Errorf("invalid domain: %s", domain)
	}

	for i := range parts[0:] {
		potentialTLD := strings.Join(parts[i:], ".")
		// fmt.Println(potentialTLD)
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
