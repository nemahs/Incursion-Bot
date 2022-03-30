package Chat

type MessageType int

const (
	PrivateMessage MessageType = iota
	ChannelMessage
	Unknown
)

type ChatMsg struct {
	Sender string
	Type   MessageType
	Text   string
}

type ChatServer interface {
	BroadcastToChannel(message string, channel string) error
	BroadcastToDefaultChannel(message string) error
	ReplyToMsg(message string, origMsg ChatMsg) error
	SendToUser(message string, user string) error
	GetNextChatMessage() (ChatMsg, error)
}