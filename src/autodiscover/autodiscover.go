package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// 为方便解析Response中的xml文件，定义结构体映射Autodiscover Response xml结构
type AutodiscoverResponse struct {
	XMLName  xml.Name `xml:"Autodiscover"`
	Response Response `xml:"Response"`
}
type Response struct {
	XMLName xml.Name `xml:"Response"`
	User    User     `xml:"User"`
	Account Account  `xml:"Account"`
	Error   *Error   `xml:"Error,omitempty"`
}

type User struct {
	AutoDiscoverSMTPAddress string `xml:"AutoDiscoverSMTPAddress"`
	DisplayName             string `xml:"DisplayName"`
	LegacyDN                string `xml:"LegacyDN"`
	DeploymentId            string `xml:"DeploymentId"`
} //DefaultABView absent

type Account struct {
	XMLName         xml.Name `xml:"Account"`
	AccountType     string   `xml:"AccountType"`
	Action          string   `xml:"Action"` //重点
	MicrosoftOnline string   `xml:"MicrosoftOnline"`
	ConsumerMailbox string   `xml:"ConsumerMailbox"`
	//AlternativeMailbox []AlternativeMailbox `xml:"AlternativeMailbox"`
	Protocol     Protocol `xml:"Protocol"` //...
	RedirectAddr string   `xml:"RedirectAddr"`
	RedirectUrl  string   `xml:"RedirectUrl"`
} //...

type Protocol struct {
}

type Error struct {
	XMLName   xml.Name `xml:"Error"`
	Time      string   `xml:"Time,attr"`
	Id        string   `xml:"Id,attr"`
	DebugData string   `xml:"DebugData"`
	Errorcode string   `xml:"Errorcode"`
	Message   string   `xml:"Message"`
}

func main() {
	domain := "outlook.com"             //域名，但实际上还需要考虑客户端使用的是自己生成的域名的情况
	email_address := "info@outlook.com" //客户端需要配置的邮件地址

	//通过[MS-OXDISCO]中的3.1.5指出的方法找到Autodiscover server的URI
	//var uris []string;  //声明需要的server uris切片
	//1.Query a well-known LDAP server for service connection point objects(暂忽略)

	//2.Perform text manipulations on the domain of the email address(对给定的邮件地址文本直接进行操作)
	uris := []string{
		fmt.Sprintf("http://%s/Autodiscover/Autodiscover.xml", domain),
		fmt.Sprintf("https://autodiscover.%s/Autodiscover/Autodiscover.xml", domain), //HTTPS?
	}
	//If an HTTP POST to either of the above URIs results in an HTTP 302 redirect
	//首先遍历上述uris,发送HTTP POST请求，看是否导致302重定向
	for _, uri := range uris {
		fmt.Println("Attempting URI:", uri)
		config, err := getAutodiscoverConfig(uri, email_address)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Config: %s\n", config)
		}
		fmt.Println("\n")
	}

	//3.An Autodiscover client can query DNS to obtain SRV records for the Autodiscover service
	service := "_autodiscover._tcp." + domain
	//查询SRV记录
	_, srvs, err := net.LookupSRV("autodiscover", "tcp", domain)
	if err != nil {
		fmt.Printf("Failed to lookup SRV records for %s: %v\n", service, err)
		return
	}
	//构建URIs列表
	var uris_dns []string
	for _, srv := range srvs {
		host := strings.Trim(srv.Target, ".")
		uri_dns := fmt.Sprintf("https://%s/Autodiscover/Autodiscover.xml", host)
		uris_dns = append(uris_dns, uri_dns)
	}
	if len(uris_dns) == 0 {
		fmt.Println("No valid SRV records found.")
	} else {
		fmt.Println("Possible Autodiscover URIs:")
		for _, uri_dns := range uris_dns {
			fmt.Println(uri_dns)
			config, err := getAutodiscoverConfig(uri_dns, email_address)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Config: %s\n", config)
			}
			fmt.Println("\n")
		}
	}

	//4.An Autodiscover client can also issue an HTTP GET method with the URI set to "http://Autodiscover.<domain>/Autodiscover/Autodiscover.xml"

}

// getAutodiscoverConfig函数实现发送HTTP POST请求以及检查是否导致302重定向
func getAutodiscoverConfig(uri string, email_add string) (string, error) {
	xmlRequest := fmt.Sprintf(`
		<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
            <Request>
                <EMailAddress>%s</EMailAddress>
                <AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
            </Request>
        </Autodiscover>`, email_add) //按照[MS-OSDSCLI]4.1的格式

	resp, err := http.Post(uri, "text/xml", bytes.NewBufferString(xmlRequest)) //发送HTTP POST请求
	//fmt.Println("00")
	if err != nil {
		fmt.Println("09")
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("Response Status Code: %d\n", resp.StatusCode)

	//fmt.Println("1")
	if resp.StatusCode == http.StatusFound { //if HTTP 302
		//fmt.Println("22")
		//the client SHOULD repost the request to the redirection URL contained in the Location header
		redirect_uri := resp.Header.Get("Location")           //从Response的Location Header中提取重定向uri进行repost
		fmt.Printf("Redirecting to: %s\n", redirect_uri)      //[MS-OSDSCLI]3.1.5.2 HTTP 302 Redirects
		return getAutodiscoverConfig(redirect_uri, email_add) //Repost
	} else if resp.StatusCode == http.StatusOK { //如果Autodiscover Server返回了Autodiscover Response [MS-OSDSCLI]3.2.5
		//fmt.Println("22")
		//解析Autodiscover Response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response body: %v\n", err)
			return "", fmt.Errorf("failed to read response body: %v", err)
		}
		//fmt.Println("22")
		var autodiscoverResp AutodiscoverResponse
		err = xml.Unmarshal(body, &autodiscoverResp)
		if err != nil {
			fmt.Printf("Failed to unmarshal XML: %v\n", err)
			return "", fmt.Errorf("failed to unmarshal XML: %v", err)
		}

		/*If the server returns an Autodiscover response (as specified in section 2.2.4)
		which contains an Action element (section 2.2.4.1.1.2.2) with a value of "redirectAddr",
		the client SHOULD send a new Autodiscover request.*/
		//检查Action类型
		//[MS-OSDSCLI]3.1.5.3 Autodiscover Redirect
		//若为redirectAddr
		if autodiscoverResp.Response.Account.Action == "redirectAddr" {
			//fmt.Println("22")
			newEmail := autodiscoverResp.Response.Account.RedirectAddr
			if newEmail != "" {
				fmt.Printf("RedirectAddr: %s\n", newEmail) //可打印
				//重新发送请求
				return getAutodiscoverConfig(uri, newEmail)
			}
		} else if autodiscoverResp.Response.Account.Action == "redirectUrl" { //若为redirectUrl
			//fmt.Println("22")
			newUri := autodiscoverResp.Response.Account.RedirectUrl
			if newUri != "" {
				fmt.Printf("RedirectUrl: %s\n", newUri) //可打印
				return getAutodiscoverConfig(newUri, email_add)
			}
		} else if autodiscoverResp.Response.Error != nil {
			fmt.Printf("ErrorCode: %s\n", autodiscoverResp.Response.Error.Errorcode)

		} else {
			return string(body), nil // 返回XML配置文件内容
		}

	}
	return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)

}
