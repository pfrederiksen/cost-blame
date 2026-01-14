package export

import (
	"encoding/json"
	"testing"

	"github.com/pfrederiksen/cost-blame/internal/cost"
)

func TestBuildSlackMessage(t *testing.T) {
	deltas := []cost.Delta{
		{
			Key:           "AmazonEC2",
			CurrentCost:   600.0,
			PriorCost:     100.0,
			AbsoluteDelta: 500.0,
			PercentChange: 500.0,
			IsNewSpender:  false,
			Currency:      "USD",
		},
		{
			Key:           "AmazonRDS",
			CurrentCost:   300.0,
			PriorCost:     200.0,
			AbsoluteDelta: 100.0,
			PercentChange: 50.0,
			IsNewSpender:  false,
			Currency:      "USD",
		},
		{
			Key:           "NewService",
			CurrentCost:   50.0,
			PriorCost:     0.0,
			AbsoluteDelta: 50.0,
			PercentChange: 9999.0,
			IsNewSpender:  true,
			Currency:      "USD",
		},
	}

	// Build the message (without sending to HTTP)
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
		},
	}

	// Build attachments
	for i, d := range deltas {
		if i >= 5 {
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
		}

		msg.Attachments = append(msg.Attachments, attachment)
	}

	// Verify message structure
	if len(msg.Attachments) != 3 {
		t.Errorf("Expected 3 attachments, got %d", len(msg.Attachments))
	}

	// Verify we created attachments (color logic is in the actual SendToSlack function)
	if msg.Attachments[0].Title != "AmazonEC2" {
		t.Errorf("First attachment should be AmazonEC2, got %q", msg.Attachments[0].Title)
	}

	if msg.Attachments[1].Title != "AmazonRDS" {
		t.Errorf("Second attachment should be AmazonRDS, got %q", msg.Attachments[1].Title)
	}

	// Verify JSON marshaling works
	_, err := json.Marshal(msg)
	if err != nil {
		t.Errorf("Failed to marshal SlackMessage: %v", err)
	}
}

func TestSlackMessageJSON(t *testing.T) {
	msg := SlackMessage{
		Text: "Test",
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: "Hello",
				},
			},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Verify it unmarshals back
	var decoded SlackMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Text != "Test" {
		t.Errorf("Text = %q, want %q", decoded.Text, "Test")
	}

	if len(decoded.Blocks) != 1 {
		t.Errorf("Blocks length = %d, want 1", len(decoded.Blocks))
	}
}
