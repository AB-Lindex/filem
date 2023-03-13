package message

// Sender is the abstract interface a message-sender needs to fullfill
type Sender interface {
	Close() error
	Send(buf []byte, headers map[string]string, dryRun bool) (string, error)
}
