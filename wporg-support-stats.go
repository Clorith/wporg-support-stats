package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/robfig/cron"
)

// Configuration holds a structured setup of config values.
type Configuration struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Site     string `json:"site"`
	Schedule string `json:"schedule"`
}

// Addresses holds a structured setup of URLs we will use.
type Addresses struct {
	Tags    string
	Replies string
	Topics  string
	Users   string
}

// Config holds the configuration struct
var Config Configuration

// Address holds the addresses we will be querying
var Address Addresses

var (
	wd, _ = os.Getwd()
)

func main() {
	Config = loadConfig()
	Address = loadAddress()

	fetchFromDotorg()

	c := cron.New()

	// Set up the schedule to run the stats gathering at.
	switch Config.Schedule {
	case "weekly":
		c.AddFunc("@weekly", func() { fetchFromDotorg() })
	case "daily":
		c.AddFunc("@daily", func() { fetchFromDotorg() })
	case "hourly":
		c.AddFunc("@hourly", func() { fetchFromDotorg() })
	default:
		c.AddFunc("@hourly", func() { fetchFromDotorg() })
	}

	c.Start()

	// Run indefinitely (or at least until we cancel it).
	select {}
}

func fetchFromDotorg() {
	options := cookiejar.Options{}

	jar, err := cookiejar.New(&options)

	if err != nil {
		log.Fatal(err)
	}

	client := http.Client{Jar: jar}
	resp, err := client.PostForm("https://login.wordpress.org/wp-login.php", url.Values{
		"log": {Config.Username},
		"pwd": {Config.Password},
	})

	if err != nil {
		log.Fatal(err)
	}

	// Fetch tags.
	resp, err = client.Get(Address.Tags)
	if err != nil {
		log.Fatal(err)
	}

	tags, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	tagR := regexp.MustCompile(`(?ms)"displaying-num">(.+?) items<`)
	tagResults := tagR.FindStringSubmatch(string(tags))

	if len(tagResults) > 0 {
		fmt.Printf("Total tags: %s", tagResults[1])
	}

	// Fetch replies.
	resp, err = client.Get(Address.Replies)
	if err != nil {
		log.Fatal(err)
	}

	replies, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	replyR := regexp.MustCompile(`(?ms)All.+?"count">\((.+?)\)<.+?Published.+?"count">\((.+?)\)<.+?Archived.+?"count">\((.+?)\)<`)
	replyResults := replyR.FindStringSubmatch(string(replies))

	/*
	 * 1: Total replies
	 * 2: Published replies
	 * 3: Archived replies
	 */
	if len(replyResults) > 0 {
		fmt.Printf("Total replies: %s", replyResults[1])
	}

	// Fetch topics.
	resp, err = client.Get(Address.Topics)
	if err != nil {
		log.Fatal(err)
	}

	topics, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	topicR := regexp.MustCompile(`(?ms)All.+?"count">\((.+?)\)<.+?Published.+?"count">\((.+?)\)<.+?Closed.+?"count">\((.+?)\)<.+?Archived.+?"count">\((.+?)\)<`)
	topicResults := topicR.FindStringSubmatch(string(topics))

	/*
	 * 1: All topics
	 * 2: Published topics
	 * 3: Closed topics
	 * 4: Archived topics
	 */
	if len(topicResults) > 0 {
		fmt.Printf("All topics: %s", topicResults[1])
	}

	// Fetch users.
	resp, err = client.Get(Address.Users)
	if err != nil {
		log.Fatal(err)
	}

	users, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	userR := regexp.MustCompile(`(?ms)All.+?"count">\((.+?)\)<.+?Administrator.+?"count">\((.+?)\)<.+?Editor.+?"count">\((.+?)\)<.+?Keymaster.+?"count">\((.+?)\)<.+?Moderator.+?"count">\((.+?)\)<.+?Blocked.+?"count">\((.+?)\)<.+?HelpHub Editor.+?"count">\((.+?)\)<.+?HelpHub Manager.+?"count">\((.+?)\)<`)
	userResults := userR.FindStringSubmatch(string(users))

	/*
	 * 1: Total users
	 * 2: Administrators
	 * 3: Editors
	 * 4: Keymasters
	 * 5: Moderators
	 * 6: Blocked
	 * 7: HelpHub Editor
	 * 8: HelpHub Manager
	 */
	if len(topicResults) > 0 {
		fmt.Printf("Total users: %s", userResults[1])
	}

	// Prepare our data array to be written to the CSV file.
	currentTime := time.Now()
	data := []string{
		currentTime.String(),
		tagResults[1],
		replyResults[1],
		replyResults[2],
		replyResults[3],
		topicResults[1],
		topicResults[2],
		topicResults[3],
		topicResults[4],
		userResults[1],
		userResults[2],
		userResults[3],
		userResults[4],
		userResults[5],
		userResults[7],
		userResults[8],
		userResults[6],
	}

	// Set up the CSV writers.
	filename := filepath.Join(wd, "output/stats.csv")
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	err = csvWriter.Write(data)
	if err != nil {
		panic(err)
	}

	log.Println("Done processing")
}

func loadConfig() Configuration {
	var config Configuration
	configFile, err := os.Open("config.json")

	if err != nil {
		log.Fatal(err)
	}

	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	return config
}

func loadAddress() Addresses {
	var address Addresses

	address.Tags = Config.Site + "/wp-admin/edit-tags.php?taxonomy=topic-tag&post_type=topic"
	address.Replies = Config.Site + "/wp-admin/edit.php?post_type=reply"
	address.Topics = Config.Site + "/wp-admin/edit.php?post_type=topic"
	address.Users = Config.Site + "/wp-admin/users.php?role=bbp_moderator"

	return address
}

// NewError adds an error entry to the logfile by appending to it.
func newErrorLog(message error, filename string) {
	if filename == "" {
		filename = "debug"
	}

	logfile, err := os.OpenFile(filepath.Join(wd, fmt.Sprintf("%s.log", filename)), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	defer logfile.Close()

	log.SetOutput(logfile)
	log.Println(message)
}

// New adds a custom string entry to the logfile by appending to it.
func newLog(message string, filename string) {
	if filename == "" {
		filename = "debug"
	}

	logfile, err := os.OpenFile(filepath.Join(wd, fmt.Sprintf("%s.log", filename)), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	defer logfile.Close()

	log.SetOutput(logfile)
	log.Println(message)
}
