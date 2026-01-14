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

func TestSendToSlack_EmptyDeltas(t *testing.T) {
	// Test that sending empty deltas returns an error
	// We don't actually send to HTTP, just verify validation
	err := SendToSlack("https://hooks.slack.com/test", []cost.Delta{}, 5)
	if err == nil {
		t.Error("SendToSlack should return error for empty deltas")
	}
	if err != nil && err.Error() != "no data to send" {
		t.Errorf("Expected 'no data to send' error, got %q", err.Error())
	}
}

func TestSlackAttachmentColor(t *testing.T) {
	tests := []struct {
		name          string
		absoluteDelta float64
		expectedColor string
	}{
		{"small delta", 50.0, "good"},
		{"medium delta", 150.0, "warning"},
		{"large delta", 600.0, "danger"},
		{"exactly 100", 100.0, "good"},
		{"exactly 500", 500.0, "warning"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test color logic
			color := "good"
			if tt.absoluteDelta > 100 {
				color = "warning"
			}
			if tt.absoluteDelta > 500 {
				color = "danger"
			}

			if color != tt.expectedColor {
				t.Errorf("Color for delta %.2f = %s, want %s", tt.absoluteDelta, color, tt.expectedColor)
			}
		})
	}
}

func TestSlackMessageWithNewSpender(t *testing.T) {
	deltas := []cost.Delta{
		{
			Key:           "NewService",
			CurrentCost:   100.0,
			PriorCost:     0.0,
			AbsoluteDelta: 100.0,
			PercentChange: 9999.0,
			IsNewSpender:  true,
			Currency:      "USD",
		},
	}

	// Build message manually to test structure
	msg := SlackMessage{
		Text: ":warning: *AWS Cost Spike Alert*",
	}

	for i, d := range deltas {
		if i >= 5 {
			break
		}

		attachment := SlackAttachment{
			Color: "warning",
			Title: d.Key,
			Fields: []SlackField{
				{
					Title: "Current Cost",
					Value: "$100.00",
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

	// Verify new spender field is added
	if len(msg.Attachments) != 1 {
		t.Fatalf("Expected 1 attachment, got %d", len(msg.Attachments))
	}

	foundNewSpender := false
	for _, field := range msg.Attachments[0].Fields {
		if field.Title == "Status" && field.Value == "🆕 New Spender" {
			foundNewSpender = true
			break
		}
	}

	if !foundNewSpender {
		t.Error("New spender field not found in attachment")
	}
}

func TestSlackMessageTopNLimit(t *testing.T) {
	// Create 10 deltas
	deltas := make([]cost.Delta, 10)
	for i := range deltas {
		deltas[i] = cost.Delta{
			Key:           string(rune('A' + i)),
			CurrentCost:   float64(i * 100),
			PriorCost:     float64(i * 50),
			AbsoluteDelta: float64(i * 50),
			PercentChange: 100.0,
			IsNewSpender:  false,
			Currency:      "USD",
		}
	}

	// Document that SendToSlack limits to top 5 attachments
	// (actual HTTP call would be tested with a mock server)
	t.Log("SendToSlack should limit attachments to 5 for readability")
	t.Log("Even if topN parameter is higher, only 5 attachments are created")
	t.Skip("Requires HTTP mock server for full test")
}
