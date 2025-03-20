package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/serverleader/shoutrrr/pkg/services/teams"
)

func main() {
	// Real extraId example from user
	extraId := "V2ESyij_gAljSoAPHvZoAszlpAoAXExyOl26dlf1xHEx05"

	// Test URL parsing for the new format
	testURLs := []string{
		// Valid URL with all required parts
		"teams://11111111-2222-4333-8444-555555555555@99999999-8888-4777-8666-555555555555/12345678abcdef1234567890abcdef12/aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee/" + extraId + "?host=organization.webhook.office.com",

		// Invalid URL missing extraId
		"teams://11111111-2222-4333-8444-555555555555@99999999-8888-4777-8666-555555555555/12345678abcdef1234567890abcdef12/aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee?host=organization.webhook.office.com",

		// Invalid URL missing host
		"teams://11111111-2222-4333-8444-555555555555@99999999-8888-4777-8666-555555555555/12345678abcdef1234567890abcdef12/aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee/" + extraId,
	}

	logger := log.New(os.Stderr, "VALIDATOR ", log.LstdFlags)

	for i, testURL := range testURLs {
		fmt.Printf("\nTesting URL %d: %s\n", i+1, testURL)

		// Parse URL
		parsedURL, err := url.Parse(testURL)
		if err != nil {
			fmt.Printf("  Error parsing URL: %v\n", err)
			continue
		}

		// Create a service
		service := &teams.Service{}

		// Initialize the service with the URL
		err = service.Initialize(parsedURL, logger)
		if err != nil {
			fmt.Printf("  Error initializing service: %v\n", err)
			continue
		}

		fmt.Println("  Service initialized successfully")

		// Try sending a message (note: this will attempt to actually send a message)
		if i == 0 { // Only try to send with the first valid URL
			fmt.Println("  Attempting to send a test message...")

			// Uncomment to actually send a message with a real webhook URL
			// err = service.Send("This is a test message from Shoutrrr", nil)
			// if err != nil {
			//     fmt.Printf("  Error sending message: %v\n", err)
			// } else {
			//     fmt.Println("  Message sent successfully!")
			// }
		}
	}
}
