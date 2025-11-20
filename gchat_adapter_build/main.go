package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// AlertmanagerPayload is a simplified structure to capture the key parts of the Alertmanager webhook payload.
type AlertmanagerPayload struct {
	Alerts []Alert `json:"alerts"`
	Status string  `json:"status"`
}

// Alert is a simplified structure for a single alert.
type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    string            `json:"startsAt"`
	EndsAt      string            `json:"endsAt"`
}

// GoogleChatCard is a simplified structure for a Google Chat Card Message (Text + Cards format).
type GoogleChatCard struct {
	Text    string        `json:"text"`
	CardsV2 []interface{} `json:"cardsV2,omitempty"`
}

func main() {
	// The environment variable MUST be set in the docker-compose.yml
	webhookURL := os.Getenv("GOOGLE_CHAT_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("Error: GOOGLE_CHAT_WEBHOOK_URL environment variable is not set.")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload AlertmanagerPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("Error decoding payload: %v", err)
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		status := payload.Status

		// Build the message text content
		var textBuffer bytes.Buffer
		// Determine icon based on status
		icon := "🚨"
		if status == "resolved" {
			icon = "✅"
		}
		
		textBuffer.WriteString(fmt.Sprintf("%s **Alert Status:** %s\n", icon, status))

		for _, alert := range payload.Alerts {
			// Extract necessary labels for display
			alertname := alert.Labels["alertname"]
			instance := alert.Labels["instance"] // This is the potentially missing variable
			severity := alert.Labels["severity"]
			summary := alert.Annotations["summary"]

			// --- DEBUG LOGGING ADDED HERE ---
			// Print all received labels to the server console for debugging.
			log.Printf("--- Alert Labels Check ---")
			log.Printf("Alert Name: %s", alertname)
			log.Printf("All Labels Received: %v", alert.Labels)
			log.Printf("--------------------------")
			// ---------------------------------


			textBuffer.WriteString(fmt.Sprintf("\n**Alert: %s**\n", alertname))
			textBuffer.WriteString(fmt.Sprintf("  ->Instance: `%s`\n", instance))
			textBuffer.WriteString(fmt.Sprintf("  ->Severity: %s\n", severity))
			textBuffer.WriteString(fmt.Sprintf("  ->Summary: %s\n", summary))
		}

		// Minimal card structure for Google Chat's V2 API
		chatMessage := GoogleChatCard{
			Text: textBuffer.String(),
		}

		// Send the message to Google Chat
		jsonData, _ := json.Marshal(chatMessage)
		resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error forwarding to Google Chat: %v", err)
			http.Error(w, "Error forwarding alert", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Google Chat webhook failed with status: %s", resp.Status)
			http.Error(w, "Webhook failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Alert forwarded successfully")
	})

	log.Println("Google Chat Adapter listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}