//+build unit

package middlewares_test

import (
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdhityaRamadhanus/userland/pkg/common/http/middlewares"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuthentication(t *testing.T) {
	t.Skip()
	username := "test"
	password := "coba"
	testCases := []struct {
		AuthHeader         string
		ExpectedStatusCode int
	}{
		{
			AuthHeader:         fmt.Sprintf("Basic %s", b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))),
			ExpectedStatusCode: http.StatusOK,
		},
		{
			AuthHeader:         "Bearer ngasal",
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			AuthHeader:         "Basic test",
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			AuthHeader:         fmt.Sprintf("Basic %s", b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", password, password)))),
			ExpectedStatusCode: http.StatusUnauthorized,
		},
	}

	for _, testCase := range testCases {
		authenticator := middlewares.BasicAuth(username, password)
		ts := httptest.NewServer(authenticator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})))
		defer ts.Close()

		req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
		req.Header.Set("Authorization", testCase.AuthHeader)
		res, err := http.DefaultClient.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, testCase.ExpectedStatusCode)
	}
}
