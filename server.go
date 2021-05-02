package main

import "crypto/md5"
import "encoding/hex"
import "net/http"
import "net/url"
import "fmt"
import "encoding/json"
import "io/ioutil"
import "time"
import "strconv"
import "github.com/gofiber/fiber/v2"

import "github.com/gofiber/fiber/v2/middleware/cache"
import yaml "gopkg.in/yaml.v2"

type Character struct {
	Id          int
	Name        string
	Description string
}

type Characters struct {
	Results []Character
}

type ApiResponse struct {
	Data Characters
}

type Config struct {
	PublicKey  string `yaml:"public_key"`
	PrivateKey string `yaml:"private_key"`
}

func setup(config *Config) {
	yamlFile, err := ioutil.ReadFile("env.yaml")
	if err != nil {
		panic("Credentials not found")
	}

	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		panic("Error processing yaml")
	}
}

func authQs(config *Config) url.Values {
	var ts = time.Now().Unix()
	hashStr := fmt.Sprintf("%d%s%s", ts, config.PrivateKey, config.PublicKey)
	hash := md5.Sum([]byte(hashStr))
	q := url.Values{}
	q.Add("ts", strconv.FormatInt(ts, 10))
	q.Add("apikey", config.PublicKey)
	q.Add("hash", hex.EncodeToString(hash[:]))

	return q
}

func main() {
	var config Config
	setup(&config)

	app := fiber.New()
	app.Use(cache.New(cache.Config{
		Expiration:   10 * time.Second,
		CacheControl: true,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("👋")
	})

	app.Get("/characters", func(c *fiber.Ctx) error {
		marvelUrl := "https://gateway.marvel.com:443/v1/public/characters"
		client := &http.Client{}
		request, err := http.NewRequest("GET", marvelUrl, nil)

		if err != nil {
			fmt.Printf("Request Error %v", err)
		}

		request.URL.RawQuery = authQs(&config).Encode()

		fmt.Println("qs? %s", request.URL.String())

		request.Header.Set("content-type", "application/json; charset=UTF-8")

		// 1493 characters as of may 2
		response, err := client.Do(request)

		defer response.Body.Close()

		if err != nil {
			fmt.Printf("Response Error %v", err)
		}

		body, _ := ioutil.ReadAll(response.Body)

		var data ApiResponse
		json.Unmarshal(body, &data)

		for _, character := range data.Data.Results {
			fmt.Println(">> ", character.Id, character.Name)
		}

		out, _ := json.Marshal(data.Data.Results)
		return c.SendString(fmt.Sprintf("%d %s", len(data.Data.Results), string(out)))
	})

	app.Get("/characters/:id", func(c *fiber.Ctx) error {
		marvelUrl := fmt.Sprintf("%s%s", "https://gateway.marvel.com:443/v1/public/characters/", c.Params("id"))
		client := &http.Client{}
		request, err := http.NewRequest("GET", marvelUrl, nil)

		if err != nil {
			fmt.Printf("Request Error %v", err)
		}

		request.URL.RawQuery = authQs(&config).Encode()

		request.Header.Set("content-type", "application/json; charset=UTF-8")

		fmt.Println("making request >>>>>>>>>>")
		response, err := client.Do(request)

		defer response.Body.Close()

		if err != nil {
			fmt.Printf("Response Error %v", err)
		}

		var char ApiResponse
		body, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal(body, &char)

		out, _ := json.Marshal(char.Data.Results)
		return c.SendString(string(out))

	})

	app.Listen(":3000")
}
