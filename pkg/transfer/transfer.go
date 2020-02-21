package transfer


type Transfer interface {
	InitAndConnect()
	Send(string, []byte) error
}

/*
* New client for transfer interface
*/
func NewClient(url string, subFunc MessageArrivedFunc) Transfer {
	return NewMqttClient(url, subFunc)
}
