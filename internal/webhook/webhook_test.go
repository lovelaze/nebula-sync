package webhook

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/version"
)

func TestWebhook(t *testing.T) {
	t.Run("success webhook uses success configuration", func(t *testing.T) {
		// Setup test server to verify request
		var receivedHeaders http.Header
		var receivedBody string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
			buf, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			receivedBody = string(buf)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		// Create webhook settings
		settings := &config.WebhookSettings{
			Success: config.WebhookRequest{
				URL:     ts.URL,
				Method:  "POST",
				Body:    "success-body",
				Headers: map[string]string{"X-Test": "success"},
			},
			Client: config.WebhookClient{},
		}

		client := NewClient(settings)
		err := client.triggerSuccess()
		require.NoError(t, err)

		// Verify request
		assert.Equal(t, "success-body", receivedBody)
		assert.Equal(t, "success", receivedHeaders.Get("X-Test"))
		assert.Equal(t, "nebula-sync/"+version.Version, receivedHeaders.Get("User-Agent"))
	})

	t.Run("failure webhook uses failure configuration", func(t *testing.T) {
		var receivedHeaders http.Header
		var receivedBody string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
			buf, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			receivedBody = string(buf)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		settings := &config.WebhookSettings{
			Failure: config.WebhookRequest{
				URL:     ts.URL,
				Method:  "PUT",
				Body:    "failure-body",
				Headers: map[string]string{"X-Test": "failure"},
			},
			Client: config.WebhookClient{},
		}

		client := NewClient(settings)
		err := client.triggerFailure()
		require.NoError(t, err)

		assert.Equal(t, "failure-body", receivedBody)
		assert.Equal(t, "failure", receivedHeaders.Get("X-Test"))
	})

	t.Run("empty url skips webhook", func(t *testing.T) {
		settings := &config.WebhookSettings{
			Success: config.WebhookRequest{
				URL: "",
			},
		}

		client := NewClient(settings)
		err := client.triggerSuccess()
		require.NoError(t, err)
	})

	t.Run("error on non-200 response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		settings := &config.WebhookSettings{
			Success: config.WebhookRequest{
				URL: ts.URL,
			},
		}

		client := NewClient(settings)
		err := client.triggerSuccess()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "webhook returned status 400")
	})
}
