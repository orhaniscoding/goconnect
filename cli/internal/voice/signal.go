package voice

// Signal represents a WebRTC signaling message.
type Signal struct {
	Type      string `json:"type"`      // "offer", "answer", "candidate"
	SDP       string `json:"sdp"`       // JSON encoded SDP
	Candidate string `json:"candidate"` // JSON encoded ICE candidate
	SenderID  string `json:"sender_id"`
	TargetID  string `json:"target_id"`
	NetworkID string `json:"network_id"`
}
