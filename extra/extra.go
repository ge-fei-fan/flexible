package extra

type Extra struct {
	Clipboard *Appclipboard
	Websocket *wsHub
}

func NewExtra() Extra {
	return Extra{
		Clipboard: newAppcb(),
		Websocket: newWsHub(),
	}
}
