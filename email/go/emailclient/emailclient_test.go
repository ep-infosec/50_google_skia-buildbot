package emailclient

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.skia.org/infra/go/httputils"
)

func TestClientSendWithMarkup_HappyPath(t *testing.T) {
	const expectedMessageID = "<the-actual-message-id>"
	const expected = `From: Alert Manager <alerts@skia.org>
To: someone@example.org
Subject: Alert!
Content-Type: text/html; charset=UTF-8
References: some-thread-reference
In-Reply-To: some-thread-reference

<html>
<body>
<h2>Hi!</h2>

</body>
</html>
`
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		bodyAsString := string(b)
		require.Equal(t, expected, bodyAsString)
		w.Header().Add("x-message-id", expectedMessageID)
		w.WriteHeader(http.StatusOK)
	}))
	c := NewAt(s.URL)
	c.client = httputils.NewFastTimeoutClient()

	msgID, err := c.SendWithMarkup("Alert Manager", "alerts@skia.org", []string{"someone@example.org"}, "Alert!", "", "<h2>Hi!</h2>", "some-thread-reference")
	require.NoError(t, err)
	require.Equal(t, expectedMessageID, msgID)
}

func TestClientSendWithMarkup_HTTPRequestFails_ReturnsError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	c := NewAt(s.URL)

	_, err := c.SendWithMarkup("Alert Manager", "alerts@skia.org", []string{"someone@example.org"}, "Alert!", "", "<h2>Hi!</h2>", "some-thread-reference")
	require.Contains(t, err.Error(), "Failed to send")
}

func TestClientSendWithMarkup_DedupRecipients(t *testing.T) {
	const expectedMessageID = "<the-actual-message-id>"
	const expected = `From: Alert Manager <alerts@skia.org>
To: someone@example.org,Someone Else <someone-else@example.org>
Subject: Alert!
Content-Type: text/html; charset=UTF-8
References: some-thread-reference
In-Reply-To: some-thread-reference

<html>
<body>
<h2>Hi!</h2>

</body>
</html>
`
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		bodyAsString := string(b)
		require.Equal(t, expected, bodyAsString)
		w.Header().Add("x-message-id", expectedMessageID)
		w.WriteHeader(http.StatusOK)
	}))
	c := NewAt(s.URL)
	c.client = httputils.NewFastTimeoutClient()

	msgID, err := c.SendWithMarkup("Alert Manager", "alerts@skia.org", []string{"someone@example.org", "Someone <someone@example.org>", "<someone@example.org>", "Someone Else <someone-else@example.org>"}, "Alert!", "", "<h2>Hi!</h2>", "some-thread-reference")
	require.NoError(t, err)
	require.Equal(t, expectedMessageID, msgID)
}

func TestClientValid(t *testing.T) {
	require.True(t, NewAt(DefaultEmailServiceURL).Valid())
	require.True(t, NewAt(NamespacedEmailServiceURL).Valid())
	require.False(t, NewAt("").Valid())
}
