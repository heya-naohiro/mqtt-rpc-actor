package main

import (
	"fmt"
	"io"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type mqttWriter struct {
	payload []byte
}

func (w *mqttWriter) Write(p []byte) (n int, err error) {
	w.payload = append(w.payload, p...)
	return len(p), nil
}

func (w *mqttWriter) Close() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")
	c := mqtt.NewClient(opts)
	fmt.Println(string(w.payload))
	token := c.Publish("go-mqtt/sample", 1, false, w.payload)
	token.Wait()
	c.Disconnect(250)
	fmt.Println("Complete publish")
	w.payload = make([]byte, 0)
	return nil
}

func main() {
	w := &mqttWriter{}
	f, err := os.Open("test.payload")
	if err != nil {
		fmt.Println("error")
	}
	defer f.Close()
	io.Copy(w, f)
	w.Close()

	fmt.Println("Hello")
	time.Sleep(time.Second * 1)
}
