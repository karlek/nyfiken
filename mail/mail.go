// Package mail notifies the user about updates on checked pages.
package mail

import "net/smtp"
import "net/url"

import "github.com/karlek/nyfiken/settings"
import "github.com/mewkiz/pkg/errorsutil"

// Send sends a mail to a mail address with the updates of the checked page and
// the URL to the checked page.
func Send(pageUrl url.URL, toMailAddress string, body string) (err error) {
	// Set up authentication information.
	auth := smtp.PlainAuth(
		"",
		settings.Global.SenderMail.Address,
		settings.Global.SenderMail.Password,
		settings.Global.SenderMail.AuthServer,
	)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	var msg = `From: ` + settings.Global.SenderMail.Address + `
To: ` + toMailAddress + `
Subject: [ nyfiken ] ` + pageUrl.Host + `: update
MIME-Version: 1.0
Content-Transfer-Encoding: 8bit
Content-Type: text/html; charset="UTF-8"

` + `<a href="` + pageUrl.String() + `">` + pageUrl.String() + `</a> has been updated :) <hr>
` + body + `</body><html>` + settings.Global.Newline

	err = smtp.SendMail(
		settings.Global.SenderMail.OutServer, // Outgoing server.
		auth, // Authorization information.
		settings.Global.SenderMail.Address, // From what mail.
		[]string{toMailAddress},            // To which mail.
		[]byte(msg),                        // Content to send.
	)
	if err != nil {
		return errorsutil.ErrorfColor("%s", err)
	}

	return nil
}
