package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var calendarId = "khayyam.saleem@gmail.com"

func getClient(ll *log.Logger, config *oauth2.Config) *http.Client {
	tokenFile := "token.json"
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token = getTokenFromWeb(ll, config)
		saveToken(ll, tokenFile, token)
	}
	return config.Client(context.Background(), token)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func getTokenFromWeb(ll *log.Logger, config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	ll.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		ll.Fatalf("Unable to read authorization code: %v", err)
	}

	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		ll.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return token
}

func saveToken(ll *log.Logger, path string, token *oauth2.Token) {
	f, err := os.Create(path)
	if err != nil {
		ll.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getStartDate(ll *log.Logger, recurringRule string) string {
	now := time.Now().In(time.UTC)

	// Check if the recurrence rule includes a BYDAY part
	ruleRegex := regexp.MustCompile(`BYDAY=([A-Za-z]+)`)
	matches := ruleRegex.FindStringSubmatch(recurringRule)
	if len(matches) > 0 {
		// Extract the day of the week from the BYDAY part of the rule
		dayOfWeek := matches[1][:2]

		// Find the next occurrence of the desired day of the week
		for {
			if strings.ToUpper(now.Weekday().String()[:2]) == dayOfWeek {
				return now.Format("2006-01-02")
			}
			now = now.AddDate(0, 0, 1)
		}
	} else {
		// No BYDAY part found, assume the event occurs every day
		return now.Format("2006-01-02")
	}
}

func createEvent(ll *log.Logger, srv *calendar.Service, calendarId string, event *calendar.Event) (*calendar.Event, error) {
	// Check if the event already exists
	events, err := srv.Events.List(calendarId).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve events: %v", err)
	}
	for _, item := range events.Items {
		if item.Summary == event.Summary {
			log.Printf("Event already exists: %v\n", item.Summary)
			return nil, nil
		}
	}

	// If event doesn't exist, create it
	createdEvent, err := srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		return nil, err
	}
	ll.Printf("Event created: %v\n", createdEvent.Summary)
	return createdEvent, nil
}

func main() {
	ll := log.Default()

	ll.Println("Starting...")

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		ll.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		ll.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ll, config)

	srv, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		ll.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	csvFile, err := os.Open("schedule.csv")
	if err != nil {
		ll.Fatalf("Unable to open schedule CSV file: %v", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)

	// Skip the header row
	_, err = reader.Read()
	if err != nil {
		log.Fatalf("Error reading the header row: %v", err)
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			ll.Fatalf("Error reading CSV file: %v", err)
		}

		eventName := record[0]
		startTime := record[1]
		endTime := record[2]
		recurringRule := record[3]

		startDate := getStartDate(ll, recurringRule)
		start, err := time.Parse(time.RFC3339, startDate+"T"+startTime+"-04:00")
		if err != nil {
			panic(err)
		}
		end, err := time.Parse(time.RFC3339, startDate+"T"+endTime+"-04:00")
		if err != nil {
			panic(err)
		}

		event := &calendar.Event{
			Summary: eventName,
			Start: &calendar.EventDateTime{
				DateTime: start.UTC().Format(time.RFC3339),
				TimeZone: "UTC",
			},
			End: &calendar.EventDateTime{
				DateTime: end.UTC().Format(time.RFC3339),
				TimeZone: "UTC",
			},
		}

		if recurringRule != "" {
			event.Recurrence = []string{"RRULE:" + recurringRule}
		}

		ll.Printf("Creating event: %s at time %s\n", eventName, event.Start.DateTime)

		createdEvent, err := createEvent(ll, srv, calendarId, event)
		if err != nil {
			ll.Fatalf("Unable to create event: %v\n", err)
		}

		ll.Printf("Event created: %s\n", createdEvent.HtmlLink)
	}
	ll.Printf("Done!")
}
