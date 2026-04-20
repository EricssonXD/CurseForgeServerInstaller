package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetJSON_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "hello" {
			t.Errorf("missing header")
		}
		if r.URL.Query().Get("q") != "search" {
			t.Errorf("missing param q=%s", r.URL.Query().Get("q"))
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer srv.Close()

	result, err := GetJSONWithOptions(srv.URL, map[string]string{"X-Test": "hello"}, map[string]string{"q": "search"}, 1, 0, 5*time.Second)
	if err != nil {
		t.Fatalf("GetJSON: %v", err)
	}
	if result["ok"] != true {
		t.Errorf("result = %v", result)
	}
}

func TestGetJSON_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte("forbidden"))
	}))
	defer srv.Close()

	_, err := GetJSONWithOptions(srv.URL, nil, nil, 1, 0, 5*time.Second)
	if err == nil {
		t.Fatal("expected error")
	}
	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if httpErr.StatusCode != 403 {
		t.Errorf("status = %d", httpErr.StatusCode)
	}
}

func TestGetJSON_Retries(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(500)
			w.Write([]byte("error"))
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"attempt": float64(attempts)})
	}))
	defer srv.Close()

	result, err := GetJSONWithOptions(srv.URL, nil, nil, 3, 10*time.Millisecond, 5*time.Second)
	if err != nil {
		t.Fatalf("GetJSON: %v", err)
	}
	if result["attempt"] != float64(3) {
		t.Errorf("expected attempt 3, got %v", result["attempt"])
	}
}
