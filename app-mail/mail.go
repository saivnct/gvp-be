package appmail

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"sync"
	"time"
)

type MailSendMsg struct {
	Receivers   []string
	Subject     string
	ContentType string
	Body        string
}

type MailService struct {
	SMTPHost       string
	SMTPPort       int
	SMTPUserName   string
	SMTPPassword   string
	SMTPAddress    string
	SendMsgChannel chan *MailSendMsg
}

var (
	singletonMailService *MailService
	onceMailService      sync.Once
)

func GetMailService() *MailService {
	onceMailService.Do(func() {
		fmt.Println("Init MailService...")

		singletonMailService = &MailService{}
	})
	return singletonMailService
}

func (mailService *MailService) Start(smtpHost string, smtpPort int, smtpUsername string, smtpPass string, smtpAddress string) {

	sendWSMsgChannel := make(chan *MailSendMsg, 1000) //make a buffered channel size 1000
	mailService.SMTPHost = smtpHost
	mailService.SMTPPort = smtpPort
	mailService.SMTPUserName = smtpUsername
	mailService.SMTPPassword = smtpPass
	mailService.SMTPAddress = smtpAddress
	mailService.SendMsgChannel = sendWSMsgChannel

	//test connection
	dialer := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPass)

	sendCloser, err := dialer.Dial()
	if err != nil {
		log.Fatalf("Failed to connect to mail server: %v", err)
	}

	err = sendCloser.Close()
	if err != nil {
		log.Fatalf("Failed to disconnect from mail server: %v", err)
	}

	go mailDaemon(
		mailService.SMTPHost,
		mailService.SMTPPort,
		mailService.SMTPUserName,
		mailService.SMTPPassword,
		mailService.SMTPAddress,
		mailService.SendMsgChannel)

	fmt.Println("MailService Started!")
}

func (mailService *MailService) Stop() {
	if mailService.SendMsgChannel != nil {
		close(mailService.SendMsgChannel)
	}
}

func (mailService *MailService) SendMail(receivers []string, subject string, contentType string, body string) {
	mailService.SendMsgChannel <- &MailSendMsg{
		Receivers:   receivers,
		Subject:     subject,
		ContentType: contentType,
		Body:        body,
	}
}

func mailDaemon(smtpHost string, smtpPort int, username string, password string, fromAddress string, sendMsgChannel chan *MailSendMsg) {

	dialer := gomail.NewDialer(smtpHost, smtpPort, username, password)

	var sendCloser gomail.SendCloser
	var err error
	open := false
	for {
		select {
		case mailSendMsg, ok := <-sendMsgChannel:
			if !ok {
				return
			}
			//log.Println("on sendMsgChannel")
			//log.Println(mailSendMsg.Receivers, mailSendMsg.Subject, mailSendMsg.Body)

			if !open {
				if sendCloser, err = dialer.Dial(); err != nil {
					log.Println("mailDaemon - Failed to connect to mail server", err)
					continue
				}
				open = true
			}

			message := gomail.NewMessage()
			message.SetHeader("From", fromAddress)
			message.SetHeader("To", mailSendMsg.Receivers...)
			message.SetHeader("Subject", mailSendMsg.Subject)
			//message.SetBody("text/html", mailSendMsg.Body)
			message.SetBody(mailSendMsg.ContentType, mailSendMsg.Body)

			if err := gomail.Send(sendCloser, message); err != nil {
				log.Print(err)
			}
		// Close the connection to the SMTP server if no email was sent in
		// the last 30 seconds.
		case <-time.After(30 * time.Second):
			if open {
				if err := sendCloser.Close(); err != nil {
					log.Println("mailDaemon - Failed to disconnect from mail server", err)
					continue
				}
				open = false
			}
		}
	}
}
