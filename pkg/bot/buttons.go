package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// ActionButton represents an interactive button for WhatsApp messages
type ActionButton struct {
	Name        string // "quick_reply", "cta_url", "cta_copy"
	DisplayText string
	ID          string // for quick_reply
	URL         string // for cta_url
	CopyCode    string // for cta_copy
}

// QuickReplyButton creates a quick reply button
func QuickReplyButton(displayText, id string) ActionButton {
	return ActionButton{
		Name:        "quick_reply",
		DisplayText: displayText,
		ID:          id,
	}
}

// URLButton creates a URL link button
func URLButton(displayText, url string) ActionButton {
	return ActionButton{
		Name:        "cta_url",
		DisplayText: displayText,
		URL:         url,
	}
}

// CopyButton creates a copy code button
func CopyButton(displayText, code string) ActionButton {
	return ActionButton{
		Name:        "cta_copy",
		DisplayText: displayText,
		CopyCode:    code,
	}
}

// SendInteractiveMessage sends an interactive Native Flow button message with plain text fallback
func (r *Router) SendInteractiveMessage(target types.JID, title, bodyText, footerText string, buttons []ActionButton) error {
	if r.waClient == nil || !r.waClient.IsConnected() {
		return fmt.Errorf("whatsapp client is not connected")
	}

	if len(buttons) == 0 {
		return r.SendMessage(target, bodyText)
	}

	var protoButtons []*waProto.InteractiveMessage_NativeFlowMessage_NativeFlowButton
	for _, btn := range buttons {
		var paramsMap map[string]interface{}
		switch btn.Name {
		case "cta_url":
			paramsMap = map[string]interface{}{
				"display_text": btn.DisplayText,
				"url":          btn.URL,
				"merchant_url": btn.URL,
			}
		case "cta_copy":
			paramsMap = map[string]interface{}{
				"display_text": btn.DisplayText,
				"id":           btn.CopyCode,
				"copy_code":    btn.CopyCode,
			}
		default: // "quick_reply"
			paramsMap = map[string]interface{}{
				"display_text": btn.DisplayText,
				"id":           btn.ID,
			}
		}

		paramsBytes, err := json.Marshal(paramsMap)
		if err != nil {
			continue
		}

		protoButtons = append(protoButtons, &waProto.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
			Name:             proto.String(btn.Name),
			ButtonParamsJSON: proto.String(string(paramsBytes)),
		})
	}

	interactiveMsg := &waProto.InteractiveMessage{
		Body: &waProto.InteractiveMessage_Body{
			Text: proto.String(bodyText),
		},
		InteractiveMessage: &waProto.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waProto.InteractiveMessage_NativeFlowMessage{
				Buttons: protoButtons,
			},
		},
	}

	if title != "" {
		interactiveMsg.Header = &waProto.InteractiveMessage_Header{
			Title: proto.String(title),
		}
	}
	if footerText != "" {
		interactiveMsg.Footer = &waProto.InteractiveMessage_Footer{
			Text: proto.String(footerText),
		}
	}

	msg := &waProto.Message{
		ViewOnceMessage: &waProto.FutureProofMessage{
			Message: &waProto.Message{
				InteractiveMessage: interactiveMsg,
			},
		},
	}

	_, err := r.waClient.SendMessage(context.Background(), target, msg)
	if err != nil {
		log.Printf("[Bot] Interactive message failed (%v), falling back to plain text for %s", err, target.String())
		return r.SendMessage(target, bodyText)
	}

	return nil
}
