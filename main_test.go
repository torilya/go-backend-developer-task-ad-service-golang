package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

type testCase struct {
	id             int
	data           any
	expectedResult any
}

// Надо еще проверять файлы bidRequests/new.json и new.log
type testAdHandlerResult struct {
	statusCode int
	body       string
}

func TestAdHandler(t *testing.T) {
	//goland:noinspection LongLine
	testCases := []testCase{
		{
			id: 1,
			data: url.Values{
				"client": []string{"1"},
				"slot":   []string{"1"},
				"user":   []string{"1"},
			},
			expectedResult: testAdHandlerResult{
				statusCode: http.StatusOK,
				body:       "<a href=\"http://ya.ru/\"><img src=\"http://via.placeholder.com/600x600\" width=\"600\" height=\"600\" border=\"0\" alt=\"Advertisement\" /></a>",
			},
		},
	}

	for _, tc := range testCases {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet,
			"http://example.com/ad?"+tc.data.(url.Values).Encode(),
			nil)

		adHandler(recorder, request)

		expectedResult := tc.expectedResult.(testAdHandlerResult)

		response := recorder.Result()

		body, _ := io.ReadAll(response.Body)

		//defer response.Body.Close()

		actualResult := testAdHandlerResult{
			statusCode: recorder.Code,
			body:       string(body),
		}

		if !reflect.DeepEqual(expectedResult, actualResult) {
			t.Errorf("[%d] Wrong result:\n\tExpected:\t%+v\n\tActual:\t\t%+v", tc.id,
				expectedResult, actualResult)
		}
	}
}
