package appgrpc

import (
	"fmt"
	appmail "gbb.go/gvp/app-mail"
	"gbb.go/gvp/model"
	"log"
)

const (
	Email_Authencode_Title      = "Signup confirmation"
	Email_Authencode_ContenType = "text/plain"
)

func (sv *XVPGRPCService) SendEmailAuthenCode(userTmp *model.UserTmp) error {
	log.Println("Call SendSMSAuthenCode")

	body := fmt.Sprintf("Hello, This is you Signup confirmation code: %v", userTmp.Authencode)

	mailSv := appmail.GetMailService()
	mailSv.SendMail([]string{userTmp.Email}, Email_Authencode_Title, Email_Authencode_ContenType, body)
	return nil
}
