package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func splitEmailAddress(emailAddress string) []string {
	parts := strings.Split(emailAddress, "@")
	//fmt.Println(parts)
	return parts
}

// described in 4.3 MX use nslookup or dig to get the MX record
func get_nslookup_(domain string) []string {
	out, err := exec.Command("nslookup -q=mx ", domain).Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Println(string(out))
	lines := strings.Split(string(out), "\n")
	//fmt.Println(lines)
	var mx_records []string
	for _, line := range lines {
		if strings.Contains(line, "mail exchanger") {
			mx_records = append(mx_records, line)
		}
	}
	//fmt.Println(mx_records)
	return mx_records
}

func get_url_autoconfig(email_address string) {
	parts := splitEmailAddress(email_address)
	email_domain := parts[1]
	email_local := parts[0]

	// 1.1. https://autoconfig.%EMAILDOMAIN%/mail/config-v1.1.xml?emailaddress=%EMAILADDRESS% (Required)
	url_1_1 := "https://autoconfig." + email_domain + "/mail/config-v1.1.xml?emailaddress=" + email_address

	// 1.2. https://%EMAILDOMAIN%/.well-known/autoconfig/mail/config-v1.1.xml (Recommended)
	url_1_2 := "https://" + email_domain + "/.well-known/autoconfig/mail/config-v1.1.xml"

	// 1.3. http://autoconfig.%EMAILDOMAIN%/mail/config-v1.1.xml(Optional)
	url_1_3 := "http://autoconfig." + email_domain + "/mail/config-v1.1.xml"

	// 2.1. %ISPDB%%EMAILDOMAIN% (Recommended)
	// %ISPDB% = https://autoconfig.thunderbird.net/v1.1/
	url_2_1 := "https://autoconfig.thunderbird.net/v1.1/" + email_domain

	// 3

}
