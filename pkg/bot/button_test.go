package bot

import (
	"encoding/json"
	"testing"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func TestExtractButtonResponses(t *testing.T) {
	// Test 1: ButtonsResponseMessage
	msg1 := &waProto.Message{
		ButtonsResponseMessage: &waProto.ButtonsResponseMessage{
			SelectedButtonID: proto.String("!bimbingan"),
		},
	}
	if text := extractMessageText(msg1); text != "!bimbingan" {
		t.Errorf("expected '!bimbingan', got '%s'", text)
	}

	// Test 2: ListResponseMessage
	msg2 := &waProto.Message{
		ListResponseMessage: &waProto.ListResponseMessage{
			SingleSelectReply: &waProto.ListResponseMessage_SingleSelectReply{
				SelectedRowID: proto.String("!jadwal"),
			},
		},
	}
	if text := extractMessageText(msg2); text != "!jadwal" {
		t.Errorf("expected '!jadwal', got '%s'", text)
	}

	// Test 3: TemplateButtonReplyMessage
	msg3 := &waProto.Message{
		TemplateButtonReplyMessage: &waProto.TemplateButtonReplyMessage{
			SelectedID: proto.String("!acc 1 0"),
		},
	}
	if text := extractMessageText(msg3); text != "!acc 1 0" {
		t.Errorf("expected '!acc 1 0', got '%s'", text)
	}

	// Test 4: InteractiveResponseMessage
	msg4 := &waProto.Message{
		InteractiveResponseMessage: &waProto.InteractiveResponseMessage{
			InteractiveResponseMessage: &waProto.InteractiveResponseMessage_NativeFlowResponseMessage_{
				NativeFlowResponseMessage: &waProto.InteractiveResponseMessage_NativeFlowResponseMessage{
					ParamsJSON: proto.String(`{"id":"!revisi 1 1"}`),
				},
			},
		},
	}
	if text := extractMessageText(msg4); text != "!revisi 1 1" {
		t.Errorf("expected '!revisi 1 1', got '%s'", text)
	}
}

func TestBuildInteractiveButton(t *testing.T) {
	paramsJSON, _ := json.Marshal(map[string]string{
		"display_text": "📋 Bimbingan",
		"id":           "!bimbingan",
	})

	btn := &waProto.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
		Name:             proto.String("quick_reply"),
		ButtonParamsJSON: proto.String(string(paramsJSON)),
	}

	msg := &waProto.Message{
		ViewOnceMessage: &waProto.FutureProofMessage{
			Message: &waProto.Message{
				InteractiveMessage: &waProto.InteractiveMessage{
					Header: &waProto.InteractiveMessage_Header{
						Title: proto.String("Menu Utama"),
					},
					Body: &waProto.InteractiveMessage_Body{
						Text: proto.String("Pilih menu di bawah ini:"),
					},
					Footer: &waProto.InteractiveMessage_Footer{
						Text: proto.String("Laboratorium ASCII"),
					},
					InteractiveMessage: &waProto.InteractiveMessage_NativeFlowMessage_{
						NativeFlowMessage: &waProto.InteractiveMessage_NativeFlowMessage{
							Buttons: []*waProto.InteractiveMessage_NativeFlowMessage_NativeFlowButton{btn},
						},
					},
				},
			},
		},
	}

	if msg.ViewOnceMessage.Message.InteractiveMessage.Header.GetTitle() != "Menu Utama" {
		t.Errorf("expected title 'Menu Utama'")
	}
}
