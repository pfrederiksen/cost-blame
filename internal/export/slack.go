package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pfrederiksen/cost-blame/internal/cost"
)

// SlackMessage represents a Slack webhook payload
type SlackMessage struct {
	Text        string             `json:"text,omitempty"`
	Blocks      []SlackBlock       `json:"blocks,omitempty"`
	Attachments []SlackAttachment  `json:"attachments,omitempty"`
}

// SlackBlock represents a Slack block
type SlackBlock struct {
	Type string          `json:"type"`
	Text *SlackText      `json:"text,omitempty"`
	Fields []SlackText   `json:"fields,omitempty"`
}

// SlackText represents Slack text
type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SlackAttachment represents a Slack attachment
type SlackAttachment struct {
	Color  string `json:"color,omitempty"`
	Title  string `json:"title,omitempty"`
	Text   string `json:"text,omitempty"`
	Fields []SlackField `json:"fields,omitempty"`
}

// SlackField represents a Slack field
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// SendToSlack sends cost spike data to a Slack webhook
func SendToSlack(webhookURL string, deltas []cost.Delta, topN int) error {
	if len(deltas) == 0 {
		return fmt.Errorf("no data to send")
	}

	// Limit to top N
	if topN > 0 && len(deltas) > topN {
		deltas = deltas[:topN]
	}

	// Build message
	msg := SlackMessage{
		Text: ":warning: *AWS Cost Spike Alert*",
		Blocks: []SlackBlock{
			{
				Type: "header",
				Text: &SlackText{
					Type: "plain_text",
					Text: "AWS Cost Spike Detected",
				},
			},
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: fmt.Sprintf("Detected %d cost changes. Top movers:", len(deltas)),
				},
			},
		},
	}

	// Add top spikes as attachments
	for i, d := range deltas {
		if i >= 5 { // Limit to 5 for readability
			break
		}

		color := "good"
		if d.AbsoluteDelta > 100 {
			color = "warning"
		}
		if d.AbsoluteDelta > 500 {
			color = "danger"
		}

		attachment := SlackAttachment{
			Color: color,
			Title: d.Key,
			Fields: []SlackField{
				{
					Title: "Current Cost",
					Value: fmt.Sprintf("$%.2f", d.CurrentCost),
					Short: true,
				},
				{
					Title: "Delta",
					Value: fmt.Sprintf("$%.2f", d.AbsoluteDelta),
					Short: true,
				},
				{
					Title: "Change",
					Value: fmt.Sprintf("%.1f%%", d.PercentChange),
					Short: true,
				},
				{
					Title: "Prior Cost",
					Value: fmt.Sprintf("$%.2f", d.PriorCost),
					Short: true,
				},
			},
		}

		if d.IsNewSpender {
			attachment.Fields = append(attachment.Fields, SlackField{
				Title: "Status",
				Value: "🆕 New Spender",
				Short: false,
			})
		}

		msg.Attachments = append(msg.Attachments, attachment)
	}

	// Send to Slack
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send to Slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}
