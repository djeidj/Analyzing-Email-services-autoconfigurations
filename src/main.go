package main

import (
	"fmt"
	"github.com/djeidj/Analyzing-Email-services-autoconfigurations/autoconfig"
	"github.com/djeidj/Analyzing-Email-services-autoconfigurations/utils"
)

func main() {
	// test for splitEmailAddress
	emailAddress := "ghffdjyuf@hvmhv.com"
	parts := utils.SplitEmailAddress(emailAddress)
	fmt.Println(parts)

}
