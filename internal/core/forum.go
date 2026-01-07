package core

type ForumChannelWithLastMessages struct {
	Channel      ForumChannel
	LastMessages []ForumMessageWithUser
}

type ForumMessageWithUser struct {
	Message ForumMessage
	User    *ForumUser
}

type ForumChannelMessagesResponse struct {
	ChannelID string
	Messages  []ForumMessageWithUser
	HasMore   bool
}
