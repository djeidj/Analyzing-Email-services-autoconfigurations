package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// 解析电子邮件地址，提取出本地部分和域名部分
func splitEmailAddress(emailAddress string) (string, string) {
	parts := strings.Split(emailAddress, "@")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// 查询并解析SRV记录
func lookupAndPrintSRV(service, proto, domain string) ([]*net.SRV, error) {
	cname, srvs, err := net.LookupSRV(service, proto, domain)
	if err != nil {
		fmt.Printf("Failed to lookup SRV record for _%s._%s.%s: %v\n", service, proto, domain, err)
		return nil, err
	}

	if len(srvs) == 0 {
		fmt.Printf("No SRV records found for _%s._%s.%s\n", service, proto, domain)
		return nil, nil
	}

	fmt.Printf("CNAME: %s\n", cname)
	for _, srv := range srvs {
		fmt.Printf("Service: _%s._%s.%s - Target: %s, Port: %d, Priority: %d, Weight: %d\n", service, proto, domain, srv.Target, srv.Port, srv.Priority, srv.Weight)
	}
	return srvs, nil
}

// 通用查询SRV记录
func lookupGeneralSRV(domain string) {
	cname, srvs, err := net.LookupSRV("", "", domain)
	if err != nil {
		fmt.Printf("Failed to lookup general SRV record for %s: %v\n", domain, err)
		return
	}

	if len(srvs) == 0 {
		fmt.Printf("No general SRV records found for %s\n", domain)
		return
	}

	fmt.Printf("CNAME: %s\n", cname)
	for _, srv := range srvs {
		fmt.Printf("General SRV record - Target: %s, Port: %d, Priority: %d, Weight: %d\n", srv.Target, srv.Port, srv.Priority, srv.Weight)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <email>")
		return
	}

	emailAddress := os.Args[1]
	_, domain := splitEmailAddress(emailAddress)
	if domain == "" {
		fmt.Println("Invalid email address")
		return
	}

	fmt.Println("Looking up SRV records for domain:", domain)
	fmt.Println()

	allFailed := true

	// 查询 _submission._tcp 的 SRV 记录
	fmt.Println("Checking for _submission._tcp SRV record...")
	if _, err := lookupAndPrintSRV("submission", "tcp", domain); err == nil {
		allFailed = false
	}
	fmt.Println()

	// 查询 _imap._tcp 的 SRV 记录
	fmt.Println("Checking for _imap._tcp SRV record...")
	if _, err := lookupAndPrintSRV("imap", "tcp", domain); err == nil {
		allFailed = false
	}
	fmt.Println()

	// 查询 _pop3._tcp 的 SRV 记录
	fmt.Println("Checking for _pop3._tcp SRV record...")
	if _, err := lookupAndPrintSRV("pop3", "tcp", domain); err == nil {
		allFailed = false
	}
	fmt.Println()

	if allFailed {
		fmt.Println("All specific SRV record lookups failed. Attempting general SRV lookup...")
		lookupGeneralSRV(domain)
	}
}
