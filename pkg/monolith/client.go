package monolith

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net"
	"reflect"
)

var proxies = make(map[reflect.Type]func(i Instance) any)

func RegisterProxy[T any](create func(i Instance) any) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	proxies[t] = create
}

type Client struct {
	name              string
	requestRoutes     SyncMap[string, *gob.Encoder]
	responseRoutes    SyncMap[string, chan response]
	address           *net.TCPAddr
	dispatcherAddress *net.TCPAddr
	logger            *log.Logger
}

func NewClient(name, endPoint, dispatcherEndPoint string) (client Client, err error) {
	address, err := net.ResolveTCPAddr("tcp", endPoint)
	if err != nil {
		return
	}
	dispatcherAddress, err := net.ResolveTCPAddr("tcp", dispatcherEndPoint)
	if err != nil {
		return
	}
	client = Client{
		name:              name,
		requestRoutes:     NewSyncMap[string, *gob.Encoder](),
		responseRoutes:    NewSyncMap[string, chan response](),
		address:           address,
		dispatcherAddress: dispatcherAddress,
		logger:            log.Default(),
	}
	return
}

func (c *Client) logf(format string, v ...any) {
	if c.logger == nil {
		return
	}
	message := fmt.Sprintf(format, v...)
	c.logger.Printf("Client '%v': %v\n", c.name, message)
}

func (c *Client) log(v ...any) {
	if c.logger == nil {
		return
	}
	message := fmt.Sprint(v...)
	c.logger.Printf("Client '%v': %v\n", c.name, message)
}

func Get[T any](id string, client *Client) (proxy T, err error) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	i := Instance{
		ID:     id,
		Type:   t.Name(),
		client: client,
	}
	p, ok := proxies[t]
	if !ok {
		err = ProxyTypeNotFoundError
		return
	}
	return p(i).(T), nil
}

func (i Instance) Call(method string, params any, results any) (err error) {
	var buffer bytes.Buffer
	err = gob.NewEncoder(&buffer).Encode(params)
	if err != nil {
		return
	}
	req := request{
		ID:       uuid.NewString(),
		Instance: i,
		Method:   method,
		Params:   buffer.Bytes(),
	}
	res := i.send(req)
	if res.Err != nil {
		return res.Err
	}
	return gob.NewDecoder(bytes.NewBuffer(res.Results)).Decode(results)
}

func (i Instance) send(req request) (res response) {
	encoder, ok := i.client.requestRoutes.get(i.Type)
	if !ok {
		encoder, res.Err = i.connect()
		if res.Err != nil {
			return
		}
	}
	route := make(chan response)
	i.client.responseRoutes.put(req.ID, route)
	defer i.client.responseRoutes.delete(req.ID)
	res.Err = encoder.Encode(req)
	if res.Err != nil {
		return
	}
	i.client.log("sent request with ID ", req.ID)
	res = <-route
	return
}

func (i Instance) connect() (encoder *gob.Encoder, err error) {
	endPoint, err := i.getEndPoint()
	if err != nil {
		return
	}
	if endPoint == "" {
		return nil, ServiceNotFoundError
	}
	i.client.logf("received endpoint %v for service '%v'", endPoint, i.Type)
	serviceAddress, err := net.ResolveTCPAddr("tcp", endPoint)
	if err != nil {
		return
	}
	conn, err := net.DialTCP("tcp", i.client.address, serviceAddress)
	if err != nil {
		return
	}
	remote := conn.RemoteAddr().String()
	i.client.log("connected to server ", remote)
	encoder = gob.NewEncoder(conn)
	i.client.requestRoutes.put(i.Type, encoder)
	go func() {
		defer func() {
			err := conn.Close()
			if err != nil {
				i.client.log(err)
			}
			i.client.requestRoutes.delete(i.Type)
		}()
		decoder := gob.NewDecoder(conn)
		for {
			var res response
			err := decoder.Decode(&res)
			if err != nil {
				i.client.log(err)
				return
			}
			i.client.log("received response with ID ", res.ID, " from server ", remote)
			route, ok := i.client.responseRoutes.get(res.ID)
			if !ok {
				i.client.log(RequestNotFoundError)
				continue
			}
			route <- res
		}
	}()
	return
}

func (i Instance) getEndPoint() (endPoint string, err error) {
	conn, err := net.DialTCP("tcp", i.client.address, i.client.dispatcherAddress)
	if err != nil {
		return
	}
	defer func() {
		closeErr := conn.Close()
		if err == nil {
			err = closeErr
		} else {
			i.client.log(closeErr)
		}
	}()
	i.client.log("connected to dispatcher ", i.client.dispatcherAddress.String())
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)
	err = encoder.Encode(i.Type)
	i.client.logf("requested endpoint for service '%v'", i.Type)
	if err != nil {
		return
	}
	err = decoder.Decode(&endPoint)
	return
}
