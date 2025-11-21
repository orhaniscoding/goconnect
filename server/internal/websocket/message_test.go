package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestDefaultMessageHandler_HandleMessage_UnknownType(t *testing.T) {
	handler := NewDefaultMessageHandler(nil, nil, nil, nil)

	client := &Client{
		userID: "user1",
		send:   make(chan []byte, 10),
	}

	msg := &InboundMessage{
		Type: "unknown.type",
		OpID: "op-1",
		Data: json.RawMessage(`{}`),
	}

	err := handler.HandleMessage(context.Background(), client, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown message type")
}

func TestDefaultMessageHandler_HandleAuthRefresh(t *testing.T) {
	handler := newTestHandler()

	client := &Client{
		userID: "user-1",
		send:   make(chan []byte, 10),
	}

	// Generate valid refresh token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "user-1",
		"user_id": "user-1",
		"type":    "refresh",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	secret := []byte("dev-secret-change-in-production")
	tokenString, _ := token.SignedString(secret)

	msg := &InboundMessage{
		Type: TypeAuthRefresh,
		OpID: "op-1",
		Data: json.RawMessage(fmt.Sprintf(`{"refresh_token":"%s"}`, tokenString)),
	}

	err := handler.HandleMessage(context.Background(), client, msg)
	assert.NoError(t, err)

	// Should receive ACK
	select {
	case data := <-client.send:
		var received OutboundMessage
		err := json.Unmarshal(data, &received)
		assert.NoError(t, err)
		assert.Equal(t, TypeAck, received.Type)
		assert.Equal(t, "op-1", received.OpID)
	default:
		t.Fatal("no ACK received")
	}
}

func TestChatSendData_Unmarshal(t *testing.T) {
	jsonData := `{
		"scope": "network:123",
		"body": "Test message",
		"attachments": ["file1.jpg", "file2.png"]
	}`

	var data ChatSendData
	err := json.Unmarshal([]byte(jsonData), &data)

	assert.NoError(t, err)
	assert.Equal(t, "network:123", data.Scope)
	assert.Equal(t, "Test message", data.Body)
	assert.Len(t, data.Attachments, 2)
	assert.Equal(t, "file1.jpg", data.Attachments[0])
}

func TestChatEditData_Unmarshal(t *testing.T) {
	jsonData := `{
		"message_id": "msg-123",
		"new_body": "Updated message"
	}`

	var data ChatEditData
	err := json.Unmarshal([]byte(jsonData), &data)

	assert.NoError(t, err)
	assert.Equal(t, "msg-123", data.MessageID)
	assert.Equal(t, "Updated message", data.NewBody)
}

func TestChatDeleteData_Unmarshal(t *testing.T) {
	jsonData := `{
		"message_id": "msg-123",
		"mode": "soft"
	}`

	var data ChatDeleteData
	err := json.Unmarshal([]byte(jsonData), &data)

	assert.NoError(t, err)
	assert.Equal(t, "msg-123", data.MessageID)
	assert.Equal(t, "soft", data.Mode)
}

func TestChatRedactData_Unmarshal(t *testing.T) {
	jsonData := `{
		"message_id": "msg-123",
		"mask": "***"
	}`

	var data ChatRedactData
	err := json.Unmarshal([]byte(jsonData), &data)

	assert.NoError(t, err)
	assert.Equal(t, "msg-123", data.MessageID)
	assert.Equal(t, "***", data.Mask)
}

func TestOutboundMessage_Marshal(t *testing.T) {
	msg := &OutboundMessage{
		Type: TypeChatMessage,
		OpID: "op-123",
		Data: &ChatMessageData{
			ID:     "msg-1",
			Scope:  "network:1",
			UserID: "user-1",
			Body:   "Hello world",
		},
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"type":"chat.message"`)
	assert.Contains(t, string(data), `"op_id":"op-123"`)
}

func TestOutboundMessage_Error(t *testing.T) {
	msg := &OutboundMessage{
		Type: TypeError,
		OpID: "op-456",
		Error: &ErrorData{
			Code:    "ERR_INVALID",
			Message: "Invalid request",
			Details: map[string]string{
				"field": "scope",
			},
		},
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"type":"error"`)
	assert.Contains(t, string(data), `"ERR_INVALID"`)
	assert.Contains(t, string(data), `"Invalid request"`)
}

func TestInboundMessage_Unmarshal(t *testing.T) {
	jsonData := `{
		"type": "chat.send",
		"op_id": "op-789",
		"data": {
			"scope": "network:123",
			"body": "Test"
		}
	}`

	var msg InboundMessage
	err := json.Unmarshal([]byte(jsonData), &msg)

	assert.NoError(t, err)
	assert.Equal(t, TypeChatSend, msg.Type)
	assert.Equal(t, "op-789", msg.OpID)

	// Parse nested data
	var sendData ChatSendData
	err = json.Unmarshal(msg.Data, &sendData)
	assert.NoError(t, err)
	assert.Equal(t, "network:123", sendData.Scope)
	assert.Equal(t, "Test", sendData.Body)
}

func TestMessageTypes_Constants(t *testing.T) {
	// Verify all message type constants are defined correctly
	assert.Equal(t, MessageType("auth.refresh"), TypeAuthRefresh)
	assert.Equal(t, MessageType("chat.send"), TypeChatSend)
	assert.Equal(t, MessageType("chat.edit"), TypeChatEdit)
	assert.Equal(t, MessageType("chat.delete"), TypeChatDelete)
	assert.Equal(t, MessageType("chat.redact"), TypeChatRedact)

	assert.Equal(t, MessageType("chat.message"), TypeChatMessage)
	assert.Equal(t, MessageType("chat.edited"), TypeChatEdited)
	assert.Equal(t, MessageType("chat.deleted"), TypeChatDeleted)
	assert.Equal(t, MessageType("chat.redacted"), TypeChatRedacted)

	assert.Equal(t, MessageType("member.joined"), TypeMemberJoined)
	assert.Equal(t, MessageType("member.left"), TypeMemberLeft)
	assert.Equal(t, MessageType("request.join.pending"), TypeJoinPending)

	assert.Equal(t, MessageType("admin.kick"), TypeAdminKick)
	assert.Equal(t, MessageType("admin.ban"), TypeAdminBan)
	assert.Equal(t, MessageType("net.updated"), TypeNetUpdated)

	assert.Equal(t, MessageType("error"), TypeError)
	assert.Equal(t, MessageType("ack"), TypeAck)
}

func TestMemberEventData_Marshal(t *testing.T) {
	data := &MemberEventData{
		NetworkID: "net-123",
		UserID:    "user-456",
		Role:      "admin",
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"network_id":"net-123"`)
	assert.Contains(t, string(jsonData), `"user_id":"user-456"`)
	assert.Contains(t, string(jsonData), `"role":"admin"`)
}

func TestJoinPendingData_Marshal(t *testing.T) {
	data := &JoinPendingData{
		NetworkID: "net-123",
		UserID:    "user-456",
		RequestID: "req-789",
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"network_id":"net-123"`)
	assert.Contains(t, string(jsonData), `"request_id":"req-789"`)
}

func TestAdminActionData_Marshal(t *testing.T) {
	data := &AdminActionData{
		NetworkID: "net-123",
		UserID:    "user-456",
		Reason:    "Spam",
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"network_id":"net-123"`)
	assert.Contains(t, string(jsonData), `"reason":"Spam"`)
}

func TestNetUpdatedData_Marshal(t *testing.T) {
	data := &NetUpdatedData{
		NetworkID: "net-123",
		Changes:   []string{"name", "description"},
		UpdatedBy: "user-456",
		Properties: map[string]interface{}{
			"name": "New Network Name",
		},
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"network_id":"net-123"`)
	assert.Contains(t, string(jsonData), `"updated_by":"user-456"`)
	assert.Contains(t, string(jsonData), `"changes"`)
}

func TestChatMessageData_MarshalWithDeletedAt(t *testing.T) {
	deletedAt := "2024-01-15T10:30:00Z"
	data := &ChatMessageData{
		ID:        "msg-1",
		Scope:     "network:1",
		UserID:    "user-1",
		Body:      "Test",
		Redacted:  false,
		DeletedAt: &deletedAt,
		CreatedAt: "2024-01-15T10:00:00Z",
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"deleted_at":"2024-01-15T10:30:00Z"`)
}

func TestChatMessageData_MarshalWithoutDeletedAt(t *testing.T) {
	data := &ChatMessageData{
		ID:        "msg-1",
		Scope:     "network:1",
		UserID:    "user-1",
		Body:      "Test",
		Redacted:  false,
		DeletedAt: nil,
		CreatedAt: "2024-01-15T10:00:00Z",
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	// deleted_at should be omitted
	assert.NotContains(t, string(jsonData), `"deleted_at"`)
}

func TestChatMessageData_WithAttachments(t *testing.T) {
	data := &ChatMessageData{
		ID:          "msg-1",
		Scope:       "network:1",
		UserID:      "user-1",
		Body:        "Check these files",
		Redacted:    false,
		Attachments: []string{"file1.jpg", "file2.pdf"},
		CreatedAt:   "2024-01-15T10:00:00Z",
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"attachments"`)
	assert.Contains(t, string(jsonData), `"file1.jpg"`)
	assert.Contains(t, string(jsonData), `"file2.pdf"`)
}

func TestChatMessageData_RedactedFlag(t *testing.T) {
	data := &ChatMessageData{
		ID:        "msg-1",
		Scope:     "network:1",
		UserID:    "user-1",
		Body:      "***",
		Redacted:  true,
		CreatedAt: "2024-01-15T10:00:00Z",
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"redacted":true`)
}

func TestErrorData_WithDetails(t *testing.T) {
	data := &ErrorData{
		Code:    "ERR_VALIDATION",
		Message: "Validation failed",
		Details: map[string]string{
			"field": "body",
			"error": "too_short",
		},
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"code":"ERR_VALIDATION"`)
	assert.Contains(t, string(jsonData), `"message":"Validation failed"`)
	assert.Contains(t, string(jsonData), `"details"`)
	assert.Contains(t, string(jsonData), `"field":"body"`)
}

func TestErrorData_WithoutDetails(t *testing.T) {
	data := &ErrorData{
		Code:    "ERR_GENERIC",
		Message: "Something went wrong",
		Details: nil,
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"code":"ERR_GENERIC"`)
	// details should be omitted when nil
	assert.NotContains(t, string(jsonData), `"details"`)
}

func TestChatSendData_EmptyAttachments(t *testing.T) {
	jsonData := `{
		"scope": "network:123",
		"body": "Test message"
	}`

	var data ChatSendData
	err := json.Unmarshal([]byte(jsonData), &data)

	assert.NoError(t, err)
	assert.Equal(t, "network:123", data.Scope)
	assert.Equal(t, "Test message", data.Body)
	assert.Nil(t, data.Attachments)
}

func TestChatSendData_FullyPopulated(t *testing.T) {
	data := ChatSendData{
		Scope:       "host",
		Body:        "Full message",
		Attachments: []string{"doc.pdf", "image.png", "video.mp4"},
	}

	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), `"scope":"host"`)
	assert.Contains(t, string(jsonData), `"body":"Full message"`)
	assert.Contains(t, string(jsonData), `"attachments"`)
	assert.Contains(t, string(jsonData), `"doc.pdf"`)
}

func TestOutboundMessage_MultipleTypes(t *testing.T) {
	testCases := []struct {
		name    string
		msgType MessageType
	}{
		{"ChatMessage", TypeChatMessage},
		{"ChatEdited", TypeChatEdited},
		{"ChatDeleted", TypeChatDeleted},
		{"ChatRedacted", TypeChatRedacted},
		{"MemberJoined", TypeMemberJoined},
		{"MemberLeft", TypeMemberLeft},
		{"JoinPending", TypeJoinPending},
		{"AdminKick", TypeAdminKick},
		{"AdminBan", TypeAdminBan},
		{"NetUpdated", TypeNetUpdated},
		{"Error", TypeError},
		{"Ack", TypeAck},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := &OutboundMessage{
				Type: tc.msgType,
				Data: map[string]string{"test": "data"},
			}

			jsonData, err := json.Marshal(msg)
			assert.NoError(t, err)
			assert.Contains(t, string(jsonData), string(tc.msgType))
		})
	}
}

func TestInboundMessage_AllTypes(t *testing.T) {
	testCases := []struct {
		name    string
		msgType MessageType
	}{
		{"AuthRefresh", TypeAuthRefresh},
		{"ChatSend", TypeChatSend},
		{"ChatEdit", TypeChatEdit},
		{"ChatDelete", TypeChatDelete},
		{"ChatRedact", TypeChatRedact},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonStr := `{"type":"` + string(tc.msgType) + `","op_id":"op-1","data":{}}`

			var msg InboundMessage
			err := json.Unmarshal([]byte(jsonStr), &msg)

			assert.NoError(t, err)
			assert.Equal(t, tc.msgType, msg.Type)
			assert.Equal(t, "op-1", msg.OpID)
		})
	}
}
