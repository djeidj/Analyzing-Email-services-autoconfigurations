package main

import (
	"fmt"
	"github.com/djeidj/Analyzing-Email-services-autoconfigurations/autoconfig"
	"github.com/djeidj/Analyzing-Email-services-autoconfigurations/autodiscover"
	// "github.com/djeidj/Analyzing-Email-services-autoconfigurations/utils"
)

func main() {
	// download_sufffixlist()
	// err := autoconfig.Get_PublicSuffixList("../download/public_suffix_list.josn")
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("Get_PublicSuffixList success")
	// }

	err := autoconfig.Get_UrlListAutoconfigXML("1397798409@qq.com", "../download/public_suffix_list.josn", "../download/autoconfig")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Get_UrlListAutoconfigXML success")
	}

	err = autodiscover.Download_UrlListAutodiscoverXML("kyc98409@example.com", "../download/autodiscover")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Download_UrlListAutodiscoverXML success")
	}
}
