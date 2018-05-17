package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/CyCoreSystems/cycore-web/db"
	"github.com/CyCoreSystems/sendinblue"
	"github.com/labstack/echo"
)

var contactEmailT *template.Template

func init() {
	contactEmailT = template.Must(template.New("contactEmail").Parse(contactEmailTemplate))
}

// ContactRequest describes the parameters of a contact request
type ContactRequest struct {
	Name  string `json:"name" form:"name" query:"name"`
	Email string `json:"email" form:"email" query:"email"`
}

// contactRequest handles a customer contact request
func contactRequest(c echo.Context) (err error) {

	cc := c.(*Context)

	req := new(ContactRequest)
	if err = cc.Bind(req); err != nil {
		cc.Log.Warnf("failed to parse input: %s", err.Error())
		return c.JSON(http.StatusBadRequest, NewError(errors.New("failed to read request")))
	}

	cc.Log.Debugf(`received contact request from "%s" <%s> (%s)`, req.Name, req.Email, c.RealIP())
	if err = db.LogContact(req.Name, req.Email); err != nil {
		cc.Log.Warn("failed to write contact record to database")
	}

	if req.Name == "" {
		cc.Log.Warn("empty name")
		return c.JSON(http.StatusBadRequest, NewError(errors.New("please supply a name")))
	}
	if req.Email == "" {
		cc.Log.Warn("empty email")
		return c.JSON(http.StatusBadRequest, NewError(errors.New("please supply an email")))
	}

	emailBody, err := renderContactEmail(req.Name, req.Email)
	if err != nil {
		cc.Log.Errorf("failed to render email body: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, NewError(errors.New("internal error; please retry")))
	}

	msg := &sendinblue.Message{
		Sender: &sendinblue.Address{
			Name:  "CyCore Systems, Inc",
			Email: "sys@cycoresys.com",
		},
		To:          getEmailContacts(),
		Subject:     "Contact Request",
		HTMLContent: emailBody,
		Tags:        []string{"contact-request"},
	}

	if err = msg.Send(os.Getenv("SENDINBLUE_APIKEY")); err != nil {
		cc.Log.Errorf("failed to send contact email: %s", err.Error())

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "   ")
		enc.Encode(msg) // nolint

		return c.JSON(http.StatusBadGateway, NewError(errors.New("internal error; please retry")))
	}

	cc.Log.Debug("contact request sent")
	return nil
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

func getEmailContacts() []*sendinblue.Address {

	var ret []*sendinblue.Address
	if err := json.Unmarshal([]byte(os.Getenv("CONTACT_RECIPIENTS")), &ret); err != nil {

		// Fall back to default if we fail to load from environment
		ret = append(ret, &sendinblue.Address{
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
