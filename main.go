package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	mqtt "github.com/mochi-mqtt/server/v2"

	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

type Args struct{}

type TimeServer int64

func (t *TimeServer) GiveServerTime(args *Args, reply *int64) error {
	*reply = time.Now().Unix()
	return nil
}

type RPCServer struct {
	url string
}

func (rs *RPCServer) callback(client pahomqtt.Client, msg pahomqtt.Message) {
	fmt.Println("callback is caleled")
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(rs.url)
	c := pahomqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("connect error %s\n", token.Error())
	}
	reply := time.Now().Unix()
	convertedInt64 := strconv.FormatInt(reply, 10)
	token := c.Publish("mqttrpc/b", 1, false, convertedInt64)
	if token.Error() != nil {
		fmt.Println("Publish Error &s", token.Error())
	}
	token.Wait()
}

func (rs *RPCServer) Serve() error {
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(rs.url)
	c := pahomqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("connect error %s\n", token.Error())
		return token.Error()

	}
	if subscribeToken := c.Subscribe("mqttrpc/a", 1, rs.callback); subscribeToken.Wait() && subscribeToken.Error() != nil {
		fmt.Printf("token error %s\n", subscribeToken.Error())
		return subscribeToken.Error()
	}
	return nil
}

func New(url string) *RPCServer {
	return &RPCServer{url}
}

func init() {
	// Create the new MQTT Server.
	opt := &mqtt.Options{InlineClient: true}
	server := mqtt.New(opt)
	ch := make(chan interface{})
	// Allow all connections.
	_ = server.AddHook(new(auth.AllowHook), nil)

	// Create a TCP listener on a standard port.
	tcp := listeners.NewTCP("t1", ":1883", nil)
	err := server.AddListener(tcp)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
		close(ch)
	}()
	<-ch

	callbackFn := func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
		server.Log.Info("inline client received message from subscription", "client", cl.ID, "subscriptionId", sub.Identifier, "topic", pk.TopicName, "payload", string(pk.Payload))
	}
	go func() {
		err := server.Subscribe("mqttrpc", 1, callbackFn)
		if err != nil {
			fmt.Println("subscribe error, ", err)
		}
	}()
	/* rpc server  */
	url := "tcp://localhost:1883"
	rpc_server := New(url)
	go func() {
		err := rpc_server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(time.Second * 1)
}

func main() {
	/* rpc client */
	fmt.Println("- Hello -")
	go func() {
		recv()
	}()
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")
	c := pahomqtt.NewClient(opts)

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Mqtt error: %s", token.Error())
	}

	for i := 0; i < 5; i++ {
		fmt.Println(i)
		time.Sleep(time.Second * 1)
		text := fmt.Sprintf("this is msg #%d!", i)
		token := c.Publish("mqttrpc/a", 0, false, text)
		token.Wait()
	}

	c.Disconnect(250)
}

func recv() error {
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")
	c := pahomqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("connect error %s\n", token.Error())
		return token.Error()

	}
	var f pahomqtt.MessageHandler = func(client pahomqtt.Client, msg pahomqtt.Message) {
		fmt.Println(string(msg.Payload()))
	}
	if subscribeToken := c.Subscribe("mqttrpc/b", 1, f); subscribeToken.Wait() && subscribeToken.Error() != nil {
		fmt.Printf("token error %s\n", subscribeToken.Error())
		return subscribeToken.Error()
	}
	return nil
}
