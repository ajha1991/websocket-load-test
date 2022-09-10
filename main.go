package main

import (
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

var clientMap map[string]int

var tickChannel chan string

func main() {
	log.Println("app started")

	defer checkDataInMap()
	clientMap = make(map[string]int)
	tickChannel = make(chan string, 100000)
	go checkClientConnected()
	wg := &sync.WaitGroup{}
	go updateTickCount()

	for i := 0; i < 1; i++ {
		wg.Add(1)
		go startConnLoad(wg, i)
	}
	wg.Wait()
}

func startConnLoad(wg *sync.WaitGroup, id int) {
	defer wg.Done()
	clientWG := &sync.WaitGroup{}
	for i := 0; i < 1; i++ {
		clientWG.Add(1)
		cur := time.Now()
		log.Printf("connection %d started", i+1)
		createConnection(clientWG)
		log.Printf("time taken for connection %d is %d millis", i+1, time.Since(cur).Microseconds()/1000)
	}
	fmt.Printf("all connections created in routine %d", id)
	clientWG.Wait()
}

func checkDataInMap() {
	max := math.MinInt
	min := math.MaxInt
	for _, v := range clientMap {
		if max < v {
			max = v
		}

		if min > v {
			min = v
		}
	}
	fmt.Printf("max value %d min value %d total connection %d", max, min, len(clientMap))
}

func createConnection(wg *sync.WaitGroup) error {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 30 * time.Second
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	headers := http.Header{}
	headers.Add("device", "ankur")

	cur := time.Now()

	conn, _, err := dialer.Dial("ws://192.168.1.2:8080", headers)

	log.Printf("time taken for connection %d millis", time.Since(cur).Microseconds()/1000)
	if err != nil {
		fmt.Printf("\nerror while connection %v", err)
		return err
	}
	conn.SetCloseHandler(onClose)
	uid, _ := uuid.NewUUID()
	clientId := uid.String()
	wg.Add(1)
	go readMessage(conn, clientId, wg)
	cur = time.Now()
	conn.WriteMessage(websocket.BinaryMessage, []byte("hello"))
	log.Printf("time taken for message  %d millis", time.Since(cur).Microseconds()/1000)

	return nil
}

func readMessage(conn *websocket.Conn, clientId string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		mType, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Errorf("Error while reading data %v", err)
			return
		}

		//Parsing binary data
		if mType == websocket.BinaryMessage {
			tickChannel <- clientId

		} else if mType == websocket.TextMessage {

		}
	}
}

func onClose(code int, text string) error {
	fmt.Printf("connection closed with code %d and text %s", code, text)
	return nil
}

func updateTickCount() {
	for val := range tickChannel {
		if _, exist := clientMap[val]; !exist {
			clientMap[val] = 0
		}
		clientMap[val]++
	}
}

func checkClientConnected() {
	ticker := time.NewTicker(10 * time.Second)
	done := make(chan bool)

	for {
		select {
		case <-done:
			return
		case _ = <-ticker.C:
			log.Printf("clients connected %d", len(clientMap))
		}
	}
}
