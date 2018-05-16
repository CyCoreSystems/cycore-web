package controllers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/CyCoreSystems/cycore-web/app/routes"
	"github.com/pkg/errors"
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

// Request handles a customer contact request
func (c Contact) Request(name, email string) revel.Result {
	emailBody, err := renderContactEmail(name, email)
	if err != nil {
		return c.Controller.RenderError(errors.Wrap(err, "failed to render email for contact request"))
	}

	body, err := emailRequestBody(emailBody)
	if err != nil {
		return c.Controller.RenderError(errors.Wrap(err, "failed to encode email request"))
	}

	// Don't sent when in dev mode
	if revel.Config.BoolDefault("mode.dev", true) {
		revel.INFO.Println("email:", name, email, bytes.NewBuffer(body).String())
		c.Flash.Success("request faked")
		return c.Redirect(routes.App.Index())
	}

	mreq, err := http.NewRequest("POST", "https://api.sendinblue.com/v2.0/email", bytes.NewReader(body))
	mreq.Header.Add("api-key", os.Getenv("SENDINBLUE_APIKEY"))
	mreq.Header.Add("Content-Type", "application/json")
	mreq.Header.Add("X-Mailin-Tag", "contact-request")

	resp, err := http.DefaultClient.Do(mreq)
	if err != nil {
		revel.ERROR.Println("Failed to send contact request email", name, email, err)
		return c.Controller.RenderError(errors.Wrap(err, "failed to send contact request email"))
	}
	if resp.StatusCode > 299 {
		revel.ERROR.Println("Contact request email was rejected:", name, email, resp.Status, bytes.NewBuffer(body).String())
		c.Flash.Error("failed to send context request")
		return c.Redirect(routes.App.Index())
	}

	c.Flash.Success("Request sent")
	return c.Redirect(routes.App.Index())
}

func renderContactEmail(name, email string) (string, error) {
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

func emailRequestBody(email string) ([]byte, error) {
	body := struct {
		To      map[string]string `json:"to"`
		Subject string            `json:"subject"`
		From    []string          `json:"from"`
		HTML    string            `json:"html"`
	}{
		To: map[string]string{
			"scm@cycoresys.com": "Sean C McCord",
			"ll@cycoresys.com":  "Laurel Lawson",
		},
		Subject: "Contact Request",
		From:    []string{"sys@cycoresys.com", "CyCore Systems Inc"},
		HTML:    email,
	}
	return json.Marshal(&body)
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
