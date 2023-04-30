package main

import (
	"fmt"
)

func init() {
	fmt.Println("email initt")

}
func main() {

	err := Send(&SendInput{
		Template: TemplateWelcome,
		TemplateVars: map[string]string{
			"FirstName":    "Abidh Muhsin",
			"CompanyName":  "Dev Company",
			"ContactEmail": "contact@abid.dev",
			"UserName":     "abidhmuhsin",
			"LoginUrl":     "https://abid.dev",
		},
		To: "abidhmuhsin@mysaas.com",
	})
	fmt.Println(err)

	err = Send(&SendInput{
		Template: TemplatePasswordRecovery,
		TemplateVars: map[string]string{
			"FirstName":         "Abidh Muhsin",
			"CompanyName":       "Dev Company",
			"PasswordResetLink": "https://resetpwd",
		},
		To: "abidhmuhsin@mysaas.com",
	})
	fmt.Println(err)

}
