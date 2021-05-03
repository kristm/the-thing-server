package main

import "encoding/json"
import "net/http"
import "net/http/httptest"
import "testing"
import "time"
import "strconv"
import "github.com/gofiber/fiber/v2/utils"

func TestAuthQs(t *testing.T) {
	config := Config{
		PublicKey:  "1234",
		PrivateKey: "abcde",
	}

	may := time.Date(2021, time.May, 1, 0, 0, 0, 0, time.UTC)
	now = func() time.Time { return may }

	qs := authQs(&config)
	ts := now().Unix()

	utils.AssertEqual(t, qs.Get("apikey"), "1234", "api key")
	utils.AssertEqual(t, qs.Get("ts"), strconv.FormatInt(ts, 10), "timestamp")
	utils.AssertEqual(t, qs.Get("hash"), "7b6d03ff44b2980af6248ed4773d24d1", "hash")
}

func TestFetchCharacters(t *testing.T) {
	mockOutput := `{"data": {"results": [ {"id": 1, "name": "jump-man"}, {"id": 2, "name": "invisible boy"} ]}}`
	ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		res.Write([]byte(mockOutput))
	}))

	defer ts.Close()
	var data ApiResponse
	body, err := fetchCharacters(ts.URL, 0)
	utils.AssertEqual(t, err, nil)

	json.Unmarshal(body, &data)
	utils.AssertEqual(t, len(data.Data.Results), 2)
	utils.AssertEqual(t, data.Data.Results[0].Id, 1)
	utils.AssertEqual(t, data.Data.Results[0].Name, "jump-man")
}
