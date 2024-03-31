package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"golang.org/x/net/html"
)

type mailBox struct {
	ai            LLM
	provider      mailProvider
	name          string
	conversations []*mailConversation
}

type mailConversation struct {
	mailbox  *mailBox
	messages []mailMessage
	subject  string
}

type mailMessage struct {
	conversation *mailConversation
	from         string
	date         string
	msg          string
}

func newMailbox(provider mailProvider, ai LLM) *mailBox {
	return &mailBox{
		ai:       ai,
		provider: provider,
	}
}

func (mbox *mailBox) fetch() error {
	msgs, err := mbox.provider.fetch()
	if err != nil {
		return err
	}

	conversations := make(map[string][]*providerMessage)
	for i := range msgs {
		subject := msgs[i].header["Subject"]
		c, ok := conversations[subject]
		if !ok {
			c = []*providerMessage{}
		}
		c = append(c, &msgs[i])
		conversations[subject] = c
	}

	for k, v := range conversations {
		c := &mailConversation{mailbox: mbox, subject: k}
		for i := range v {
			//strBody := decode(v[i].message)
			s, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(v[i].message))
			strBody := stripHTML(string(s))
			if len(strBody) == 0 {
				continue
			}
			//strBody := parseMessage(string(v[i].message))
			msg := mailMessage{
				conversation: c,
				from:         v[i].header["From"],
				msg:          strBody,
				date:         v[i].header["Date"],
			}
			c.messages = append(c.messages, msg)
		}
		mbox.conversations = append(mbox.conversations, c)
	}
	/*

		if strings.Contains(sender, "calendar-notification@google.com") || strings.Contains(sender, "asana.com") || strings.Contains(sender, "mailer-daemon@googlemail.com") || strings.Contains(sender, "futurevisions.atlassian.net") {
			continue
		}

		if strings.Contains(from, "calendar-notification@google.com") || strings.Contains(from, "asana.com") || strings.Contains(from, "mailer-daemon@googlemail.com") || strings.Contains(from, "futurevisions.atlassian.net") {
			continue
		}

	*/
	return nil
}

func extractTextPart(part string) string {
	parsed := ""
	headerEnd := strings.Index(part, "\n\n")
	if headerEnd == -1 {
		fmt.Println("No header found.")
		return ""
	}

	headers := part[:headerEnd]
	body := part[headerEnd+2:]

	if strings.Contains(headers, "Content-Type: text/plain") {
		if strings.Contains(headers, "Content-Transfer-Encoding: base64") {
			decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(body))
			if err != nil {
				fmt.Println("Error decoding base64 content:", err)
				return ""
			}
			body = string(decoded)
		}
		parsed += strings.TrimSpace(body)
	}
	return parsed
}

func parseMessage(rawEmail string) string {
	parsed := ""

	// Extract the boundary directly from the raw email content
	boundaryIndex := strings.Index(rawEmail, "boundary=")
	if boundaryIndex == -1 {
		fmt.Println("No boundary found.")
		return ""
	}

	boundaryStart := boundaryIndex + len("boundary=")
	boundaryEnd := strings.Index(rawEmail[boundaryStart:], "\"")
	if boundaryEnd == -1 {
		fmt.Println("Malformed boundary.")
		return ""
	}

	boundary := rawEmail[boundaryStart : boundaryStart+boundaryEnd]
	boundary = strings.Trim(boundary, "\"")

	parts := strings.Split(rawEmail, "--"+boundary)
	for _, part := range parts {
		parsed += extractTextPart(part)
	}
	return parsed
}

func (mbox *mailBox) summarize(cb func(from string, date string, subject string, message string, original string)) {
	for i := range mbox.conversations {
		for j := range mbox.conversations[i].messages {
			msg := mbox.conversations[i].messages[j]
			summary := msg.summary()
			cb(msg.from, msg.date, mbox.conversations[i].subject, summary, msg.msg)
		}
	}
}

func (m *mailMessage) removeHistory() *mailMessage {
	return nil
}

func (m *mailMessage) summary() string {
	ai := m.conversation.mailbox.ai
	msg, err := ai.summary(m.msg)
	if err != nil {
		return fmt.Sprintf("(error: %v)", err)
	}
	return msg
}

func (m *mailMessage) actionItems() []string {
	return nil
}

func decode(rawEmail string) string {
	fmt.Printf("RAW EMAIL: %v\n", rawEmail)
	// Find and extract the boundary string
	boundaryPrefix := `boundary="`
	start := strings.Index(rawEmail, boundaryPrefix)
	if start == -1 {
		fmt.Println("Boundary string not found")
		return ""
	}
	start += len(boundaryPrefix)
	end := strings.Index(rawEmail[start:], `"`)
	if end == -1 {
		fmt.Println("Invalid boundary string")
		return ""
	}
	boundary := rawEmail[start : start+end]

	// Split the email into parts based on the boundary
	parts := strings.Split(rawEmail, boundary)

	// Extract the base64 encoded part
	base64Encoded := ""
	for _, part := range parts {
		if strings.Contains(part, "Content-Transfer-Encoding: base64") {
			lines := strings.Split(part, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if !strings.HasPrefix(line, "Content-") && line != "" {
					base64Encoded += line
				}
			}
		}
	}

	// Decode the base64 content
	decoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		fmt.Println("Error decoding base64 content:", err)
		return ""
	}

	fmt.Println("Decoded content:")
	return string(decoded)
}

// stripHTML removes HTML tags and CSS styles and returns plain text
func stripHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
		return ""
	}
	var b strings.Builder
	traverser := &htmlTraverser{skip: false}
	traverser.walkNodes(doc, &b)
	return b.String()
}

type htmlTraverser struct {
	skip bool
}

// walkNodes traverses the HTML nodes, skips <style> tags, and extracts text
func (t *htmlTraverser) walkNodes(n *html.Node, b *strings.Builder) {
	if n.Type == html.ElementNode && n.Data == "style" {
		t.skip = true
	} else if n.Type == html.ElementNode && n.Data == "/style" {
		t.skip = false
	}

	if !t.skip && n.Type == html.TextNode {
		b.WriteString(n.Data)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		t.walkNodes(c, b)
	}
}
