package email

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
	"github.com/instill-ai/component/base"
	"google.golang.org/protobuf/types/known/structpb"
)

// Decide it temporarily
const EmailReadingDefaultCapacity = 100

type ReadEmailsInput struct {
	Search Search `json:"search"`
}

type Search struct {
	ServerAddress      string `json:"server-address"`
	ServerPort         int    `json:"server-port"`
	Mailbox            string `json:"mailbox"`
	SearchSubjectText  string `json:"search-subject-text,omitempty"`
	SearchFromEmail    string `json:"search-from-email,omitempty"`
	SearchToEmail      string `json:"search-to-email,omitempty"`
	Limit              int    `json:"limit,omitempty"`
	Date               string `json:"date,omitempty"`
	SearchEmailMessage string `json:"search-email-message,omitempty"`
}

type ReadEmailsOutput struct {
	Emails []Email `json:"emails"`
}

type Email struct {
	Date    string   `json:"date"`
	From    string   `json:"from"`
	To      []string `json:"to,omitempty"`
	Subject string   `json:"subject"`
	Message string   `json:"message,omitempty"`
}

func (e *execution) readEmails(input *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := ReadEmailsInput{}

	err := base.ConvertFromStructpb(input, &inputStruct)
	if err != nil {
		return nil, err
	}

	client, err := initIMAPClient(
		inputStruct.Search.ServerAddress,
		inputStruct.Search.ServerPort,
	)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	setup := e.GetSetup()
	err = client.Login(
		setup.GetFields()["email-address"].GetStringValue(),
		setup.GetFields()["password"].GetStringValue(),
	).Wait()
	if err != nil {
		return nil, err
	}

	emails, err := fetchEmails(client, inputStruct.Search)
	if err != nil {
		return nil, err
	}

	if err := client.Logout().Wait(); err != nil {
		return nil, err
	}

	outputStruct := ReadEmailsOutput{
		Emails: emails,
	}

	output, err := base.ConvertToStructpb(outputStruct)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func initIMAPClient(serverAddress string, serverPort int) (*imapclient.Client, error) {

	c, err := imapclient.DialTLS(serverAddress+":"+strconv.Itoa(serverPort), nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func fetchEmails(c *imapclient.Client, search Search) ([]Email, error) {

	mailbox := search.Mailbox
	imapMailbox := convertMailboxString(mailbox)

	selectedMbox, err := c.Select(imapMailbox, nil).Wait()
	if err != nil {
		return nil, err
	}

	emails := []Email{}

	if selectedMbox.NumMessages == 0 {
		return emails, nil
	}

	if search.Limit == 0 {
		search.Limit = EmailReadingDefaultCapacity
	}
	limit := search.Limit

	// TODO: chuang8511, Research how to fetch emails by filter and concurrency.
	// It will be done before 2024-07-26.
	for i := selectedMbox.NumMessages; limit > 0; i-- {
		limit--
		email := Email{}
		var seqSet imap.SeqSet
		seqSet.AddNum(i)

		fetchOptions := &imap.FetchOptions{
			BodySection: []*imap.FetchItemBodySection{{}},
		}
		fetchCmd := c.Fetch(seqSet, fetchOptions)
		msg := fetchCmd.Next()
		var bodySection imapclient.FetchItemDataBodySection
		var ok bool
		for {
			item := msg.Next()
			if item == nil {
				break
			}
			bodySection, ok = item.(imapclient.FetchItemDataBodySection)
			if ok {
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("FETCH command did not return body section")
		}

		mr, err := mail.CreateReader(bodySection.Literal)
		if err != nil {
			return nil, fmt.Errorf("FETCH command did not return body section")
		}

		h := mr.Header
		setEnvelope(&email, h)

		if !includeSearchCondition(email, search) {
			if err := fetchCmd.Close(); err != nil {
				return nil, err
			}
			continue
		}

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			inlineHeader, ok := p.Header.(*mail.InlineHeader)
			if ok && !isHTMLType(inlineHeader) {
				b, _ := io.ReadAll(p.Body)
				email.Message += string(b)
			}
		}

		if !includeSearchMessage(email, search) {
			if err := fetchCmd.Close(); err != nil {
				return nil, err
			}
			continue
		}

		if err := fetchCmd.Close(); err != nil {
			return nil, err
		}

		emails = append(emails, email)
	}

	return emails, nil
}

func convertMailboxString(mailbox string) string {
	switch {
	case mailbox == "INBOX":
		return "INBOX"
	case mailbox == "SENT":
		return "[Gmail]/Sent Mail"
	case mailbox == "DRAFTS":
		return "[Gmail]/Drafts"
	default:
		return "INBOX"
	}
}

func setEnvelope(email *Email, h mail.Header) {
	if date, err := h.Date(); err != nil {
		log.Fatalf("failed to get date: %v", err)
	} else {
		email.Date = date.Format(time.DateTime)
	}
	if from, err := h.AddressList("From"); err != nil {
		log.Fatalf("failed to get from: %v", err)
	} else {
		email.From = from[0].String()
	}
	if to, err := h.AddressList("To"); err != nil {
		log.Fatalf("failed to get to: %v", err)
	} else {
		email.To = []string{}
		for _, t := range to {
			email.To = append(email.To, t.String())
		}
	}
	if subject, err := h.Text("Subject"); err != nil {
		log.Fatalf("failed to get subject: %v", err)
	} else {
		email.Subject = subject
	}
}

func includeSearchCondition(email Email, search Search) bool {
	if search.SearchSubjectText != "" {
		if !strings.Contains(email.Subject, search.SearchSubjectText) {
			return false
		}
	}
	if search.SearchFromEmail != "" {
		if !strings.Contains(email.From, search.SearchFromEmail) {
			return false
		}
	}
	if search.SearchToEmail != "" {
		if !strings.Contains(strings.Join(email.To, ","), search.SearchToEmail) {
			return false
		}
	}
	if search.Date != "" {
		if !strings.Contains(email.Date, search.Date) {
			return false
		}
	}
	return true
}

func includeSearchMessage(email Email, search Search) bool {
	if search.SearchEmailMessage != "" {
		if !strings.Contains(email.Message, search.SearchEmailMessage) {
			return false
		}
	}
	return true
}

func isHTMLType(inlineHeader *mail.InlineHeader) bool {
	return strings.Contains(inlineHeader.Get("Content-Type"), "text/html")
}
