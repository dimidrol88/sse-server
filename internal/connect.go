package internal

type Connect struct {
	Id        string
	Event     string
	MessageCh chan *Message
	PingCh    chan string
	DoneCh    chan struct{}
}
