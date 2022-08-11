package socketyutils

import (
	"net/http/httptest"
	"testing"
)

func TestGetValidTopic(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com/test", nil)

	// missing topic should be marked as invalid
	_, isValid1 := GetValidTopic(w, req)
	if isValid1 {
		t.Error("Validation should have failed for request with a url without a topic")
	}

	// valid topic should be extracted
	testTopic := "sockety_pipeline"
	req = httptest.NewRequest("GET", "http://localhost:8081/subscribe?topic="+testTopic, nil)
	topic, isValid2 := GetValidTopic(w, req)

	if !isValid2 {
		t.Error("Failed to retrieve valid topic from request url")
	}

	if topic != testTopic {
		t.Errorf("Got %q but wanted %q", topic, testTopic)
	}
}
