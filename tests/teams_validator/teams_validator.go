package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/serverleader/shoutrrr/pkg/services/teams"
)

func main() {
	// Test URL parsing for the new format
	testURLs := []string{
		// Valid URL with all required parts (with UUIDs for components)
		"teams://12345678-1234-1234-1234-123456789abc@98765432-9876-9876-9876-987654321def/12345678901234567890123456789012/12345678-1234-1234-1234-123456789abc/extraId?host=organization.webhook.office.com",

		// Invalid URL missing extraId
		"teams://12345678-1234-1234-1234-123456789abc@98765432-9876-9876-9876-987654321def/12345678901234567890123456789012/12345678-1234-1234-1234-123456789abc?host=organization.webhook.office.com",

		// Invalid URL missing host
		"teams://12345678-1234-1234-1234-123456789abc@98765432-9876-9876-9876-987654321def/12345678901234567890123456789012/12345678-1234-1234-1234-123456789abc/extraId",
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

			// Uncomment to actually send a message
			// err = service.Send("This is a test message from Shoutrrr", nil)
			// if err != nil {
			//     fmt.Printf("  Error sending message: %v\n", err)
			// } else {
			//     fmt.Println("  Message sent successfully!")
			// }
		}
	}
}
