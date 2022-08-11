package main

import (
	"github.com/gorilla/websocket"
	sutils "github.com/karmakettle/websockety-mcchatface/socketyutils"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var testTopic = "sockety_pipeline"
var testJson = "{\"a\":\"json\"}"

func TestPubSubIntegration(t *testing.T) {
	// confirm that topicsAndClients map isn't already populated
	_, topicFound := topicsAndClients.Load(testTopic)
	if topicFound {
		t.Error("topicsAndClients map is already populated")
	}

	// set up mock server using the subscribe handler
	subscribeHandler := http.HandlerFunc(subscribe)
	subscribeServer := httptest.NewServer(subscribeHandler)
	defer subscribeServer.Close()

	// some hackery to make this fake server work with websockets
	serverUrl := subscribeServer.URL
	serverUrl += "/subscribe?topic=" + testTopic
	serverUrl = strings.Replace(serverUrl, "http://", "ws://", 1)
	t.Log(serverUrl)

	// subscribe two clients to the test topic
	cxn1, _, err := websocket.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		t.Error("Dial:", err)
	}
	defer cxn1.Close()

	cxn2, _, err := websocket.DefaultDialer.Dial(serverUrl, nil)
	if err != nil {
		t.Error("Dial:", err)
	}
	defer cxn2.Close()

	// lazy hack to wait for Dial() to complete
	time.Sleep(2 * time.Second)

	// assert two clients are subscribed to the test topic
	twoClients := getConnectedClients(t)
	if len(twoClients) != 2 {
		t.Errorf("Found %d client(s) subscribed to %s but expected 2", len(twoClients), testTopic)
	}

	// set up a mock server using the publish handler
	publishHandler := http.HandlerFunc(publish)
	publishServer := httptest.NewServer(publishHandler)
	defer publishServer.Close()

	publishUrl := publishServer.URL
	publishUrl += "/publish?topic=" + testTopic
	t.Log(publishUrl)

	resp := publishToTest(t, publishUrl, testJson)
	resp.Body.Close()

	// assert that both clients receive a published message
	// check first client
	cxn1Json, ok1 := sutils.ReadJson(cxn1)
	if ok1 {
		jsonString := sutils.Dump(cxn1Json)
		if !strings.Contains(jsonString, testTopic) {
			t.Errorf("Expected test topic to be present in JSON response, but instead received: %s", jsonString)
		}
	} else {
		t.Errorf("First client failed to read JSON published to %s\n", testTopic)
	}

	// check second client
	cxn2Json, ok2 := sutils.ReadJson(cxn2)
	if ok2 {
		jsonString := sutils.Dump(cxn2Json)
		if !strings.Contains(jsonString, testTopic) {
			t.Errorf("Expected test topic to be present in JSON response, but instead received: %s", jsonString)
		}
	} else {
		t.Errorf("Second client failed to read JSON published to %s\n", testTopic)
	}

	// close one of the connections to test client cleanup implementation
	cxn1.Close()

	// hack: publishing twice to the topic should remove the dead client...
	// TODO - adding ping/pong handling would take care of this cleanup
	t.Log("Publish once!")
	resp1 := publishToTest(t, publishUrl, testJson)
	resp1.Body.Close()

	t.Log("Publish twice!")
	resp2 := publishToTest(t, publishUrl, testJson)
	resp2.Body.Close()

	// assert one client left standing
	leftoverClients := getConnectedClients(t)
	if len(leftoverClients) != 1 {
		t.Errorf("Found %d client(s) subscribed to %s but expected 1", len(leftoverClients), testTopic)
	}

	// verify client doesn't exit on receiving invalid JSON
	invalidJsonResp := publishToTest(t, publishUrl, "nope")
	invalidJsonResp.Body.Close()

	// JSON error
	_, ok3 := sutils.ReadJson(cxn2)
	if ok3 {
		t.Log("Second client detected the invalid JSON, and the connection remained open")
	} else {
		t.Error("Second client failed to detect invalid JSON")
	}
}

func getConnectedClients(t *testing.T) []*websocket.Conn {
	clients, topicFound := topicsAndClients.Load(testTopic)
	var clientsSlice []*websocket.Conn

	// assert topic created / exists
	if topicFound {
		// sync.Map returns type 'any', convert to slice to enable append
		clientsSlice = clients.([]*websocket.Conn)
	} else {
		t.Errorf("%q should exist in topicsAndClients map but wasn't found", testTopic)
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
	// set up mock server with publish handler
	publishHandler := http.HandlerFunc(publish)
	publishServer := httptest.NewServer(publishHandler)
	defer publishServer.Close()

	purl := publishServer.URL
	purl += "/publish"
	t.Log(purl)

	// error response for no topic in request
	noTopicResp := publishToTest(t, purl, testJson)
	noTopicRespBody, err1 := io.ReadAll(noTopicResp.Body)
	if err1 != nil {
		t.Error(err1)
	} else if !strings.Contains(string(noTopicRespBody), "No topic specified") {
		t.Error("Request to publish with no topic should have failed")
	}
	noTopicResp.Body.Close()

	// error response on empty request data
	purl += "?topic=" + testTopic
	emptyDataResp := publishToTest(t, purl, "")
	emptyDataRespBody, err2 := io.ReadAll(emptyDataResp.Body)
	if err2 != nil {
		t.Error(err2)
	} else if !strings.Contains(string(emptyDataRespBody), "Empty body") {
		t.Error("Request to publish with no data should have failed")
	}
	emptyDataResp.Body.Close()

	// error response on invalid request json
	invalidJsonResp := publishToTest(t, purl, "nope")
	invalidJsonRespBody, err3 := io.ReadAll(invalidJsonResp.Body)
	if err3 != nil {
		t.Error(err3)
	} else if !strings.Contains(string(invalidJsonRespBody), "Invalid JSON") {
		t.Error("Request to publish with invalid JSON should have failed")
	}
	invalidJsonResp.Body.Close()

	// error response on topic that hasn't been created
	purl = strings.Replace(purl, testTopic, "newTopic", 1)
	invalidTopicResp := publishToTest(t, purl, testJson)
	invalidTopicRespBody, err4 := io.ReadAll(invalidTopicResp.Body)
	if err4 != nil {
		t.Error(err4)
	} else if !strings.Contains(string(invalidTopicRespBody), "doesn't exist") {
		t.Error("Request to publish to nonexistent topic should have failed")
	}
	invalidTopicResp.Body.Close()
}
