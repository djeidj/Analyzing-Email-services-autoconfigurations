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
func lookupSRV(service, proto, domain string) ([]*net.SRV, error) {
	_, srvs, err := net.LookupSRV(service, proto, domain)
	return srvs, err
}

// 根据优先级和权重选择SRV记录
func selectSRVRecord(srvs []*net.SRV) *net.SRV {
	if len(srvs) == 0 {
		return nil
	}

	selected := srvs[0]
	for _, srv := range srvs {
		if srv.Priority < selected.Priority || (srv.Priority == selected.Priority && srv.Weight > selected.Weight) {
			selected = srv
		}
	}
	return selected
}

// 手动输入FQDN和端口信息
func promptUserForFQDN(service string) (string, int) {
	var fqdn string
	var port int
	fmt.Printf("Please enter the FQDN for the %s service: ", service)
	fmt.Scanln(&fqdn)
	fmt.Printf("Please enter the port for the %s service: ", service)
	fmt.Scanln(&port)
	return fqdn, port
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

	services := []struct {
		Name    string
		Service string
		Proto   string
	}{
		{"SMTP Submission", "submission", "tcp"},
		{"IMAP", "imap", "tcp"},
		{"POP3", "pop3", "tcp"},
	}

	for _, s := range services {
		fmt.Printf("Looking up SRV records for %s (_%s._%s.%s):\n", s.Name, s.Service, s.Proto, domain)
		srvs, err := lookupSRV(s.Service, s.Proto, domain)
		if err != nil || len(srvs) == 0 {
			fmt.Println("No SRV records found. Prompting user for FQDN and port information.")
			fqdn, port := promptUserForFQDN(s.Name)
			fmt.Printf("Using manual configuration for %s: FQDN=%s, Port=%d\n", s.Name, fqdn, port)
			continue
		}

		selectedSRV := selectSRVRecord(srvs)
		fmt.Printf("Selected SRV record for %s: Target=%s, Port=%d, Priority=%d, Weight=%d\n", s.Name, selectedSRV.Target, selectedSRV.Port, selectedSRV.Priority, selectedSRV.Weight)
	}
}

