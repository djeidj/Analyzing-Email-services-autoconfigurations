package autodiscover

import (
	"fmt"
	"github.com/djeidj/Analyzing-Email-services-autoconfigurations/utils"
	"os"
	"path/filepath"
)

func Download_UrlListAutodiscoverXML(email_address string, path string) error {
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
	// email_local := parts[0]

	url_list := make([]string, 0)

	// 3.1.5.1
	// 目前没有实现

	// 3.1.5.2 POST maybe there exists redirect
	url_2_1 := "http://" + email_domain + "/Autodiscover/Autodiscover.xml"

	url_2_2 := "https://Autodiscover." + email_domain + "/Autodiscover/Autodiscover.xml"

	redirecturl, err := utils.Post_Autodiscoverxml(url_2_1, xmlpath)
	if redirecturl != "" && err == nil {
		url_list = append(url_list, redirecturl)
	} else if redirecturl == "" && err == nil {
		return nil
	}

	redirecturl, err = utils.Post_Autodiscoverxml(url_2_2, xmlpath)
	if redirecturl != "" && err == nil {
		if len(url_list) != 0 && url_list[0] != redirecturl {
			url_list = append(url_list, redirecturl)
		}
	} else if redirecturl == "" && err == nil {
		return nil
	}

	// 3.1.5.3
	hostname, err := utils.Get_SRV_domain(email_domain)
	if err == nil {
		url_3_1 := "https://" + hostname + "/Autodiscover/Autodiscover.xml"
		url_list = append(url_list, url_3_1)
	}

	for _, url := range url_list {
		err := utils.Get_AutodiscoverXML(url, xmlpath)
		if err == nil {
			return nil
		}
	}

	// 3.1.5.4
	url_4_1 := "http://Autodiscover." + email_domain + "/Autodiscover/Autodiscover.xml"

	redirecturl, err = utils.Get_AutodiscoverXML_redirect(url_4_1, xmlpath)
	if redirecturl != "" && err == nil {
		err = utils.Get_AutodiscoverXML(redirecturl, xmlpath)
		if err == nil {
			return nil
		}
	} else if redirecturl == "" && err == nil {
		return nil
	}

	return fmt.Errorf("can't find Autodiscoverxml file for %v", email_address)
}
