//+build unit

package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdhityaRamadhanus/userland/common/http/middlewares"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
)

func TestPanic(t *testing.T) {
	testCases := []struct {
		Handler            http.Handler
		ExpectedStatusCode int
	}{
		{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(errors.New("Panic"))
			}),
			ExpectedStatusCode: http.StatusInternalServerError,
		},
		{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}),
			ExpectedStatusCode: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		ts := httptest.NewServer(middlewares.PanicHandler(testCase.Handler))
		defer ts.Close()

		res, err := http.Get(ts.URL)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, testCase.ExpectedStatusCode)
	}
}
