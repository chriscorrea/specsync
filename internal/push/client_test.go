package push

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newListenerOnPort(port int) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
}

func TestNewSafeClient(t *testing.T) {
	t.Run("creates client with timeout", func(t *testing.T) {
		client := NewSafeClient(false)
		if client == nil {
			t.Fatal("NewSafeClient returned nil")
		}
		if client.Client == nil {
			t.Fatal("WrappedClient.Client is nil")
		}
	})

	t.Run("blocks private IP", func(t *testing.T) {
		client := NewSafeClient(false)

		// attempt to connect to private IP should fail
		_, err := client.Get("http://10.0.0.1:8080/test")
		if err == nil {
			t.Error("expected error connecting to private IP, got nil")
		}
	})

	t.Run("allows localhost when enabled", func(t *testing.T) {
		// start a local test server on port 8080
		listener, err := newListenerOnPort(8080)
		if err != nil {
			t.Skipf("could not bind to port 8080: %v", err)
		}
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		server.Listener = listener
		server.Start()
		defer server.Close()

		client := NewSafeClient(true) // allow localhost

		resp, err := client.Get(server.URL)
		if err != nil {
			t.Errorf("expected localhost to be allowed, got error: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", resp.StatusCode)
		}
	})

	t.Run("blocks localhost when disabled", func(t *testing.T) {
		client := NewSafeClient(false) // block localhost

		_, err := client.Get("http://127.0.0.1:9999/test")
		if err == nil {
			t.Error("expected error connecting to localhost when disabled, got nil")
		}
	})
}
