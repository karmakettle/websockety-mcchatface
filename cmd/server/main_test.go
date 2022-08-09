package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"net/http/httptest"
	"strings"
	"testing"
)

var testTopic = "sockety_pipeline"

func TestGetValidTopic(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "http://localhost:8081/subscribe", nil)

	// missing topic should be marked as invalid
	_, isValid := getValidTopic(w, req)
	if isValid {
		t.Error("Validation should have failed for request with a url without a topic")
	}

	testTopic := "sockety_pipeline"
	req = httptest.NewRequest("POST", "http://localhost:8081/subscribe?topic=" + testTopic, nil)
	topic, isValid := getValidTopic(w, req)
	if !isValid {
		t.Error("Failed to retrieve valid topic from request url")
	} else if topic != testTopic {
		t.Errorf("Got %q but wanted %q", topic, testTopic)
	}
}

func TestSubscribe(t *testing.T) {
	// confirm that topic / client list isn't populated for the test topic
	_, topicFound := topicsAndClients.Load(testTopic)
	if topicFound {
		t.Error("topicsAndClients map is already populated")
	}

	// set up test subscribe handler (ws endpoint)
	subscribeHandler := http.HandlerFunc(subscribe)
	subscribeServer := httptest.NewServer(subscribeHandler)
	defer subscribeServer.Close()

	surl := subscribeServer.URL
	surl += "/subscribe?topic=" + testTopic
	surl = strings.Replace(surl, "http://", "ws://", 1)
	t.Log(surl)

	// let's try subscribing a client
	cxnOne, _, err := websocket.DefaultDialer.Dial(surl, nil)
	if err != nil {
		t.Error("Dial:", err)
	}
	defer cxnOne.Close()

	clientsSlice := getClients(t)
	// assert one client subscribed
	if (len(clientsSlice) != 1) {
		t.Errorf("Found %d client(s) subscribed to %s but expected 1", len(clientsSlice), testTopic)
	}

	// subscribe another client
	cxnTwo, _, err := websocket.DefaultDialer.Dial(surl, nil)
	if err != nil {
		t.Error("Dial:", err)
	}
	defer cxnTwo.Close()

	clientsSlice = getClients(t)
	// assert two clients subscribed
	if (len(clientsSlice) != 2) {
		t.Errorf("Found %d client(s) subscribed to %s but expected 2", len(clientsSlice), testTopic)
	}

	// set up test publish handler (http endpoint)
	publishHandler := http.HandlerFunc(publish)
	publishServer := httptest.NewServer(publishHandler)
	defer publishServer.Close()

	purl := publishServer.URL
	purl += "/publish?topic=" + testTopic
	t.Log(purl)

	// closing connection and publishing twice to the topic should remove the dead client
	cxnOne.Close()
	t.Log("Publish once!")
	publishToTest(t, purl)
	t.Log("Publish twice!")
	publishToTest(t, purl)

	// assert one client left standing
	clientsSlice = getClients(t)
	if (len(clientsSlice) != 1) {
		t.Errorf("Found %d client(s) subscribed to %s but expected 1", len(clientsSlice), testTopic)
	}
}

func getClients(t *testing.T) []*websocket.Conn {
	clients, topicFound := topicsAndClients.Load(testTopic)
	var clientsSlice []*websocket.Conn
	// assert topic created / exists
	if !topicFound {
		t.Errorf("%q should exist in topicsAndClients map but wasn't found", testTopic)
	} else {
		// sync.Map returns type 'any', convert to slice to enable append
		clientsSlice = clients.([]*websocket.Conn)
	}

	return clientsSlice
}

func publishToTest(t *testing.T, url string) {
    req := httptest.NewRequest("POST", url, strings.NewReader("{\"a\":\"json\"}"))
    req.Header.Set("Content-Type", "application/json")
    // need to clear this bc of reasons, hat tip to https://stackoverflow.com/a/19607204
	req.RequestURI = ""

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error in publishToTest: %v", err)
	}
	resp.Body.Close()
}

func TestPublish(t *testing.T) {
	// error response for no topic in request
	// error response on invalid request data
	// error response on topic that hasn't been created
	// message published to all clients subscribed to a valid topic
}
