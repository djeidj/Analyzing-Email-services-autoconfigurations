package autodiscover

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Autodiscover struct
type Autodiscover struct {
	XMLName  xml.Name `xml:"Autodiscover"`
	Request  Request  `xml:"Request,omitempty"`
	Response Response `xml:"Response,omitempty"`
	XMLNS    string   `xml:"xmlns,attr"`
}

type Request struct {
	AcceptableResponseSchema string `xml:"AcceptableResponseSchema"`
	EmailAddress             string `xml:"EmailAddress,omitempty"`
	LegacyDN                 string `xml:"LegacyDN,omitempty"`
}

type Response struct {
	User    User    `xml:"User,omitempty"`
	Account Account `xml:"Account"`
	Error   Error   `xml:"Error,omitempty"`
}

type User struct {
	AutoDiscoverSMTPAddress string `xml:"AutoDiscoverSMTPAddress"`
	DefaultABView           string `xml:"DefaultABView,omitempty"`
	DeploymentId            string `xml:"DeploymentId"`
	DisplayName             string `xml:"DisplayName"`
	LegacyDN                string `xml:"LegacyDN"`
}

type Account struct {
	AccountType             string                  `xml:"AccountType,omitempty"`     // if contained, it should be "email"
	Action                  string                  `xml:"Action"`                    // it's value belongs to {"settings", "redirectUrl", "redirectAddr"}
	MicrosoftOnline         string                  `xml:"MicrosoftOnline,omitempty"` // is required when Action is "settings", The value SHOULD be "False".
	ConsumerMailbox         string                  `xml:"ConsumerMailbox,omitempty"` // is required when Action is "settings", The value SHOULD be "False".
	AlternativeMailbox      AlternativeMailbox      `xml:"AlternativeMailbox,omitempty"`
	Protocol                Protocol                `xml:"Protocol,omitempty"`
	PublicFolderInformation PublicFolderInformation `xml:"PublicFolderInformation,omitempty"`
	RedirectAddr            string                  `xml:"RedirectAddr,omitempty"`
	RedirectUrl             string                  `xml:"RedirectUrl,omitempty"`
}

type Error struct {
	Time      string `xml:"Time,attr"`
	Id        string `xml:"Id,attr"`
	DebugData string `xml:"DebugData"`
	Errorcode int    `xml:"Errorcode"`
	Message   string `xml:"Message"`
}

type AlternativeMailbox struct {
	DisplayName  string `xml:"DisplayName"`
	EmailAddress string `xml:"EmailAddress,omitempty"`
	Server       string `xml:"Server,omitempty"`
	SmtpAddress  string `xml:"SmtpAddress,omitempty"`
	Type         string `xml:"Type"` // it's value belongs to {"Archiv", "Delegate", "TeamMailbox}
}

type Protocol struct {
}

type PublicFolderInformation struct {
	SmtpAddress string `xml:"SmtpAddress"`
}

func Download_AutodiscoverXML(email_address string, path string) error {
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
	// email_local := parts[0]

	url_list := make([]string, 0)

	// MS-OXDISCO 3.1.5.1
	// 目前没有实现

	// MS-OXDISCO 3.1.5.2 POST maybe there exists redirect
	url_2_1 := "http://" + email_domain + "/Autodiscover/Autodiscover.xml"
	url_list = append(url_list, url_2_1)
	url_2_2 := "https://Autodiscover." + email_domain + "/Autodiscover/Autodiscover.xml"
	url_list = append(url_list, url_2_2)

	// MS-OXDISCO 3.1.5.3
	_, srv, err := net.LookupSRV("autodiscover", "tcp", email_domain)
	if err == nil {
		for _, s := range srv {
			url_3_1 := "https://" + strings.Trim(s.Target, ".") + "/Autodiscover/Autodiscover.xml"
			url_list = append(url_list, url_3_1)
		}
	}

	for _, url := range url_list {
		err := Post_Autodiscoverxml(url, xmlpath, email_address)
		if err == nil {
			return nil
		}
	}

	// MS-OXDISCO 3.1.5.4
	url_4_1 := "http://Autodiscover." + email_domain + "/Autodiscover/Autodiscover.xml"

	err = Get_AutodiscoverXML(url_4_1, xmlpath, email_address)
	if err == nil {
		return nil
	}

	return fmt.Errorf("can't find Autodiscoverxml file for %v", email_address)
}

func Post_Autodiscoverxml(url string, xmlpath string, email_address string) error {
	// MS-OXDSCLI 2.2.3.1.1.3 LegacyDN is not implemented
	request := Autodiscover{
		Request: Request{
			AcceptableResponseSchema: "http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a",
			EmailAddress:             email_address,
		},
		XMLNS: "http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006",
	}

	requestbyte, err := xml.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestbyte))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/xml")
	// Set Http Header
	if false {
		req.Header.Set("X-MapiHttpCapability", "1") // greater than 0
		req.Header.Set("X-AnchorMailbox", email_address)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if resp.StatusCode == http.StatusOK {

		var AD Autodiscover
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		err = xml.Unmarshal(body, &AD)
		if err != nil {
			return err
		}

		// MS-OXDSCLI 3.1.5.3
		if AD.Response.Account.RedirectAddr != "" {
			return Post_Autodiscoverxml(url, xmlpath, AD.Response.Account.RedirectAddr)
		} else if AD.Response.Account.RedirectUrl != "" {
			return Post_Autodiscoverxml(AD.Response.Account.RedirectUrl, xmlpath, email_address)
		}

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
	} else if resp.StatusCode == http.StatusFound {
		return Post_Autodiscoverxml(resp.Header.Get("Location"), xmlpath, email_address)
	}

	return fmt.Errorf("error downloading file: %v use POST", url)
}

func Get_AutodiscoverXML(url string, xmlpath string, email_address string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		outFile, err := os.Create(xmlpath)
		if err != nil {
			return fmt.Errorf("error creating file: %v", err)
		}

		defer outFile.Close()

		_, err = io.Copy(outFile, response.Body)
		if err != nil {
			return fmt.Errorf("error saving to file: %v", xmlpath)
		}

		return nil
	} else if response.StatusCode == http.StatusFound {
		return Post_Autodiscoverxml(response.Header.Get("Location"), xmlpath, email_address)
	}

	return fmt.Errorf("error downloading file: %v use GET", url)
}
