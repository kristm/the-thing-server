package main

import "crypto/md5"
import "encoding/hex"
import "net/http"
import "net/url"
import "fmt"
import "encoding/json"
import "io/ioutil"
import "time"
import "regexp"
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

// for test mocking
var now = time.Now
var config Config

func authQs(config *Config) url.Values {
	var ts = now().Unix()
	hashStr := fmt.Sprintf("%d%s%s", ts, config.PrivateKey, config.PublicKey)
	hash := md5.Sum([]byte(hashStr))
	q := url.Values{}
	q.Add("ts", strconv.FormatInt(ts, 10))
	q.Add("apikey", config.PublicKey)
	q.Add("hash", hex.EncodeToString(hash[:]))

	return q
}

func fetchCharacters(url string, offset int) ([]byte, error) {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Printf("Request Error %v", err)
		return nil, err
	}

	qs := authQs(&config)

	characterQuery, _ := regexp.MatchString("[0-9]$", url)
	if !characterQuery {
		qs.Add("limit", "100")
		qs.Add("offset", strconv.Itoa(offset))
	}

	request.URL.RawQuery = qs.Encode()
	request.Header.Set("content-type", "application/json; charset=UTF-8")

	response, err := client.Do(request)

	defer response.Body.Close()

	if err != nil {
		fmt.Printf("Response Error %v", err)
		return nil, err
	}

	body, _ := ioutil.ReadAll(response.Body)
	fmt.Printf("%s: %s\n", response.Status, request.URL.Path)
	return body, nil
}

func main() {
	setup(&config)

	app := fiber.New()
	app.Use(cache.New(cache.Config{
		Expiration:   1 * time.Hour,
		CacheControl: true,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("👋")
	})

	app.Get("/characters", func(c *fiber.Ctx) error {
		var data ApiResponse
		charIds := []int{}

		// 1493 characters as of may 2
		for i := 0; i <= 15; i++ {
			body, err := fetchCharacters("https://gateway.marvel.com:443/v1/public/characters", i*100)
			if err != nil {
				return c.SendString(err.Error())
			}
			json.Unmarshal(body, &data)
			for _, char := range data.Data.Results {
				charIds = append(charIds, char.Id)
			}
		}

		out, _ := json.Marshal(charIds)
		return c.SendString(string(out))
	})

	app.Get("/characters/:id", func(c *fiber.Ctx) error {
		var char ApiResponse
		body, err := fetchCharacters(fmt.Sprintf("%s%s", "https://gateway.marvel.com:443/v1/public/characters/", c.Params("id")), 0)

		if err != nil {
			return c.SendString(err.Error())
		}

		json.Unmarshal(body, &char)

		if len(char.Data.Results) > 0 {
			out, _ := json.Marshal(char.Data.Results[0])
			return c.SendString(string(out))
		}

		return c.SendStatus(404)
	})

	app.Listen(":8080")
}
