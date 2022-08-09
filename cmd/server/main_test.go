package main

import (
	"io"
	"net/http"
	"github.com/gorilla/websocket"
	sutils "github.com/karmakettle/websockety-mcchatface/socketyutils"
	"net/http/httptest"
	"strings"
	"testing"
)

var testTopic = "sockety_pipeline"
var testJson = "{\"a\":\"json\"}"

func TestGetValidTopic(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "http://localhost:8081/subscribe", nil)

	// missing topic should be marked as invalid
	_, isValid := sutils.GetValidTopic(w, req)
	if isValid {
		t.Error("Validation should have failed for request with a url without a topic")
	}

	testTopic := "sockety_pipeline"
	req = httptest.NewRequest("POST", "http://localhost:8081/subscribe?topic=" + testTopic, nil)
	topic, isValid := sutils.GetValidTopic(w, req)
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

	// assert that both clients receive a published message
	resp := publishToTest(t, purl, testJson)
	resp.Body.Close()
	if cxnOneJson, ok := sutils.ReadJson(cxnOne); !ok {
		t.Errorf("First client failed to read JSON published to %s\n", testTopic)
	} else {
		t.Log("JSON received! " + sutils.Dump(cxnOneJson))
	}

	if cxnTwoJson, ok := sutils.ReadJson(cxnTwo); !ok {
		t.Errorf("Second client failed to read JSON published to %s\n", testTopic)
	} else {
		t.Log("JSON received! " + sutils.Dump(cxnTwoJson))
	}

	// closing connection and publishing twice to the topic should remove the dead client
	cxnOne.Close()
	t.Log("Publish once!")
	respOne := publishToTest(t, purl, testJson)
	respOne.Body.Close()
	t.Log("Publish twice!")
	respTwo := publishToTest(t, purl, testJson)
	respTwo.Body.Close()

	// assert one client left standing
	clientsSlice = getClients(t)
	if (len(clientsSlice) != 1) {
		t.Errorf("Found %d client(s) subscribed to %s but expected 1", len(clientsSlice), testTopic)
	}

	// verify client doesn't exit on reading invalid JSON
	invalidJsonResp := publishToTest(t, purl, "nope")
	invalidJsonResp.Body.Close()

	// JSON error
	if _, ok := sutils.ReadJson(cxnTwo); ok {
		t.Log("Second client detected the invalid JSON, and the connection remained open")
	} else {
		t.Error("Second client failed to detect invalid JSON")
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

func publishToTest(t *testing.T, url string, jsonString string) *http.Response {
    req := httptest.NewRequest("POST", url, strings.NewReader(jsonString))
    req.Header.Set("Content-Type", "application/json")
    // need to clear this bc of reasons, hat tip to https://stackoverflow.com/a/19607204
	req.RequestURI = ""

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error in publishToTest: %v", err)
	}

	return resp
}

func TestPublish(t *testing.T) {
	// set up test publish handler (http endpoint)
	publishHandler := http.HandlerFunc(publish)
	publishServer := httptest.NewServer(publishHandler)
	defer publishServer.Close()

	purl := publishServer.URL
	purl += "/publish"
	t.Log(purl)

	// error response for no topic in request
	noTopicResp := publishToTest(t, purl, testJson)
	noTopicRespBody, err := io.ReadAll(noTopicResp.Body)
	if err != nil {
		t.Error(err)
	} else if !strings.Contains(string(noTopicRespBody), "No topic specified") {
		t.Error("Request to publish with no topic should have failed")
	}
	noTopicResp.Body.Close()

	// error response on empty request data
	purl += "?topic=" + testTopic
	emptyDataResp := publishToTest(t, purl, "")
	emptyDataRespBody, err := io.ReadAll(emptyDataResp.Body)
	if err != nil {
		t.Error(err)
	} else if !strings.Contains(string(emptyDataRespBody), "Empty body") {
		t.Error("Request to publish with no data should have failed")
	}
	emptyDataResp.Body.Close()

	// error response on invalid request json
	invalidJsonResp := publishToTest(t, purl, "nope")
	invalidJsonRespBody, err := io.ReadAll(invalidJsonResp.Body)
	if err != nil {
		t.Error(err)
	} else if !strings.Contains(string(invalidJsonRespBody), "Invalid JSON") {
		t.Error("Request to publish with invalid JSON should have failed")
	}
	invalidJsonResp.Body.Close()

	// error response on topic that hasn't been created
	purl = strings.Replace(purl, testTopic, "newTopic", 1)
	invalidTopicResp := publishToTest(t, purl, testJson)
	invalidTopicRespBody, err := io.ReadAll(invalidTopicResp.Body)
	if err != nil {
		t.Error(err)
	} else if !strings.Contains(string(invalidTopicRespBody), "doesn't exist") {
		t.Error("Request to publish to nonexistent topic should have failed")
	}
	invalidTopicResp.Body.Close()
}
