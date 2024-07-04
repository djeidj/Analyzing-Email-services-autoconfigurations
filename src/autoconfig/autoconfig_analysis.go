package autoconfig

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/djeidj/Analyzing-Email-services-autoconfigurations/utils"
)

func Get_UrlListAutoconfigXML(email_address string, suffixlistpath string, path string) error {
	// download the url_list's XML file to path
	xmlpath := filepath.Join(path, email_address+".xml")

	dir := filepath.Dir(xmlpath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", dir)
	}

	parts, err := utils.Split_EmailAddress(email_address)
	if err != nil {
		return err
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

	mx_full_main_domain, err := utils.Get_MX_full_main_domain(email_domain, suffixlistpath)

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
		err := utils.Get_AutoconfigXML(url, xmlpath)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("can't find Autoconfigxml file for %v", email_address)

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
