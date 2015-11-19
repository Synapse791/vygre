package main

import (
    "gopkg.in/gomail.v2"
    "fmt"
)

type VygreSMTPConfig struct {
    Host        string
    Port        int
    User        string
    Password    string
    To          string
}

func (client *VygreClient) SendInactiveNotification(image string) {
    body := `<h3>Vygre has set %s to <span style="color: #FF0000">INACTIVE</span> after 3 failed attempts at starting the container.</h3>
    <br/>
    <br/><b>What to do next?</b>
    <br/>You should connect to your host and try to run vygre in debug mode using the <b>-d</b> flag or setting the log_level to debug in the configuration file.
    <br/>If that doesn't help, try running sudo docker logs CONTAINER_ID on the failed containers and look for any errors during startup.
    <br/>Once you have debugged the issue, just restart the vygre service to reset the status of the containers.
    <br/>
    <br/>Beyond this, if you think there is an issue with the service itself, raise an issue on <a href="https://github.com/Synapse791/vygre">GitHub</a>
    <br/>
    <br/>Vygre
    `

    m := gomail.NewMessage()

    m.SetHeader("From", client.Config.SMTP.User)
    m.SetHeader("To", client.Config.SMTP.To)
    m.SetHeader("Subject", "Vygre Alert")
    m.SetBody("text/html", fmt.Sprintf(body, image))

    dialer := gomail.NewPlainDialer(client.Config.SMTP.Host, client.Config.SMTP.Port, client.Config.SMTP.User, client.Config.SMTP.Password)
    if err := dialer.DialAndSend(m); err != nil {
        client.Logger.Error(err)
    }
}

func (client *VygreClient) CheckSMTPConfig() error {
    dialer := gomail.NewPlainDialer(client.Config.SMTP.Host, client.Config.SMTP.Port, client.Config.SMTP.User, client.Config.SMTP.Password)
    sender, err := dialer.Dial()
    if err != nil { return err }
    if err := sender.Close(); err != nil { return err }
    return nil
}
