package main

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
