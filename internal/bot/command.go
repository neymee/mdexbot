package bot

type Command string

func (c Command) Endpoint() string {
	switch c {
	case CmdSubscribeBtn, CmdUnsubscribeBtn:
		return "\f" + string(c)
	case CmdText:
		return "\a" + string(c)
	default:
		return "/" + string(c)
	}
}

func (c Command) String() string {
	return string(c)
}

const (
	CmdText           Command = "text"
	CmdStart          Command = "start"
	CmdCancel         Command = "cancel"
	CmdList           Command = "list"
	CmdSubscribe      Command = "subscribe"
	CmdSubscribeBtn   Command = "subscribeBtn"
	CmdUnsubscribe    Command = "unsubscribe"
	CmdUnsubscribeBtn Command = "unsubscribeBtn"
	CmdTest           Command = "test"
)
