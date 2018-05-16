package controllers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"os"
	"time"

	"github.com/CyCoreSystems/cycore-web/app/routes"
	"github.com/CyCoreSystems/sendinblue"
	"github.com/revel/revel"
)

var contactEmailT *template.Template

func init() {
	contactEmailT = template.Must(template.New("contactEmail").Parse(contactEmailTemplate))
}

// Contact controller handles customer contact operations
type Contact struct {
	*revel.Controller
}

// ContactRequest handles a customer contact request
func (c Contact) ContactRequest(name, email string) revel.Result {

	c.Log.Info("received contact request", "name", name, "email", email, "source", c.ClientIP)

	if name == "" {
		c.Flash.Error("Please supply a name")
		return c.Redirect(routes.App.Index())
	}
	if email == "" {
		c.Flash.Error("Please supply an email address")
		return c.Redirect(routes.App.Index())
	}

	emailBody, err := c.renderContactEmail(name, email)
	if err != nil {
		c.Log.Error("failed to render contact email", "error", err)
		c.Flash.Error("Internal error encountered; please try again")
		return c.Redirect(routes.App.Index())
	}

	msg := &sendinblue.Message{
		Sender: sendinblue.Address{
			Name:  "CyCore Systems, Inc",
			Email: "sys@cycoresys.com",
		},
		To:          c.getEmailContacts(),
		Subject:     "Contact Request",
		HTMLContent: emailBody,
		Tags:        []string{"contact-request"},
	}
	if err = msg.Send(os.Getenv("SENDINBLUE_APIKEY")); err != nil {
		c.Log.Error("failed to send contact email", "error", err)
		c.Flash.Error("Request failed")
		return c.Redirect(routes.App.Index())
	}

	c.Flash.Success("Request sent")
	return c.Redirect(routes.App.Index())
}

func (c Contact) renderContactEmail(name, email string) (string, error) {
	buf := new(bytes.Buffer)
	err := contactEmailT.Execute(buf, struct {
		Name      string
		Email     string
		Timestamp string
	}{
		Name:      name,
		Email:     email,
		Timestamp: time.Now().String(),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (c Contact) getEmailContacts() []sendinblue.Address {

	var ret []sendinblue.Address
	if err := json.Unmarshal([]byte(os.Getenv("CONTACT_RECIPIENTS")), &ret); err != nil {

		// Fall back to default if we fail to load from environment
		c.Log.Warn("failed to load recipients from environment", "error", err)
		ret = append(ret, sendinblue.Address{
			Name:  "System Receiver",
			Email: "sys@cycoresys.com",
		})

	}
	return ret
}

var contactEmailTemplate = `
<html>
  <body>
    <h3>Web Contact Form</h3>
	 <ul>
	   <li><b>Name:</b> {{.Name}}</li>
	   <li><b>Email:</b> {{.Email}}</li>
	   <li><b>Timestamp:</b> {{.Timestamp}}</li>
	 </ul>
  </body>
</html>
`
