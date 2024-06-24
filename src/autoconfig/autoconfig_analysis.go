package autoconfig

import (
	"fmt"
	"github.com/djeidj/Analyzing-Email-services-autoconfigurations/utils"
)

func get_url_autoconfig(email_address string) {
	parts := utils.SplitEmailAddress(email_address)
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
