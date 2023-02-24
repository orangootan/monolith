package monolith

type Instance struct {
	Type   string
	ID     string
	client *Client
}

type request struct {
	ID       string
	Instance Instance
	Method   string
	Params   []byte
}

type response struct {
	ID      string
	Err     error
	Results []byte
}
