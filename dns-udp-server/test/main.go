package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

func main() {
	// Address of the DNS server to query.
	serverAddr := "127.0.0.1:2053"

	// Create a UDP connection.
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		log.Fatalf("Error connecting to DNS server: %v", err)
	}
	defer conn.Close()

	// List of domains to query.
	domains := []string{"example.com", "google.com", "yahoo.com", "bing.com", "github.com"}
	domainIndex := 0

	// Infinite loop to send DNS queries every 23 seconds.
	for {
		// Select a domain from the list.
		domain := domains[domainIndex]
		domainIndex = (domainIndex + 1) % len(domains) // Cycle through the list.

		// Build the DNS query for the selected domain.
		query := buildDNSQuery(domain)

		// Log the domain being queried.
		fmt.Printf("\nSending query for domain: %s\n", domain)

		// Send the query to the DNS server.
		_, err = conn.Write(query)
		if err != nil {
			log.Printf("Error sending DNS query: %v\n", err)
			continue
		}

		// Receive the response from the server.
		response := make([]byte, 512)
		n, err := conn.Read(response)
		if err != nil {
			log.Printf("Error reading DNS response: %v\n", err)
			continue
		}

		// Print the DNS response to the console.
		fmt.Printf("--- DNS Response for %s ---\n", domain)
		fmt.Printf("Received %d bytes from the server\n", n)

		// Show the DNS header.
		fmt.Println("DNS Header:")
		parseDNSHeader(response)

		// Show the DNS question section.
		fmt.Println("DNS Question:")
		parseDNSQuestion(response)

		// Show the DNS answers.
		parseDNSResponse(response[:n])

		// Wait for 3 seconds before sending the next query.
		time.Sleep(3 * time.Second)
	}
}

// Function to build a DNS query for a domain.
func buildDNSQuery(domain string) []byte {
	// Header of the DNS query.
	query := make([]byte, 12)
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Uint32() & 0xFFFF
	binary.BigEndian.PutUint16(query[0:], uint16(randomNumber)) // Unique ID.
	query[2] = 1                                                // Recursion Desired (RD) flag.
	binary.BigEndian.PutUint16(query[4:], 1)                    // Number of questions.

	// Build the question section.
	parts := strings.Split(domain, ".")
	for _, part := range parts {
		query = append(query, byte(len(part)))
		query = append(query, []byte(part)...)
	}
	query = append(query, 0)                        // End of domain name.
	query = binary.BigEndian.AppendUint16(query, 1) // Type A (IPv4).
	query = binary.BigEndian.AppendUint16(query, 1) // Class IN (Internet).

	return query
}

// Function to decode and display the DNS header.
func parseDNSHeader(response []byte) {
	if len(response) < 12 {
		fmt.Println("Invalid DNS response")
		return
	}

	// Read the header.
	id := binary.BigEndian.Uint16(response[0:])
	flags := binary.BigEndian.Uint16(response[2:])
	questionCount := binary.BigEndian.Uint16(response[4:])
	answerCount := binary.BigEndian.Uint16(response[6:])
	fmt.Printf("ID: %x, Flags: %x, Questions: %d, Answers: %d\n", id, flags, questionCount, answerCount)
}

// Function to decode and display the DNS question section.
func parseDNSQuestion(response []byte) {
	offset := 12
	for {
		// If we reach the end of the domain name.
		if response[offset] == 0 {
			offset++
			break
		}

		// Move through the domain name.
		length := int(response[offset])
		offset += length + 1
	}

	// Read the type and class.
	fmt.Printf("Type: %d, Class: %d\n", binary.BigEndian.Uint16(response[offset:]), binary.BigEndian.Uint16(response[offset+2:]))
	offset += 4
}

// Function to decode and display the DNS response.
func parseDNSResponse(response []byte) {
	if len(response) < 12 {
		fmt.Println("Invalid DNS response: insufficient size")
		return
	}

	// Read the number of answers.
	offset := 12
	answerCount := binary.BigEndian.Uint16(response[6:])
	for i := 0; i < int(answerCount); i++ {
		// Check if there is enough space in the slice before trying to read more data.
		if offset+12 > len(response) {
			fmt.Println("Invalid DNS response: not enough data to read the answer")
			return
		}

		// Skip the domain name.
		for response[offset] != 0 {
			offset += int(response[offset]) + 1
		}
		offset++

		// Read the type and class.
		if offset+4 > len(response) {
			fmt.Println("Invalid DNS response: not enough data to read type and class")
			return
		}
		recordType := binary.BigEndian.Uint16(response[offset:])
		recordClass := binary.BigEndian.Uint16(response[offset+2:])
		fmt.Printf("Type: %d, Class: %d\n", recordType, recordClass)
		offset += 4

		// Read the TTL.
		if offset+6 > len(response) {
			fmt.Println("Invalid DNS response: not enough data to read TTL")
			return
		}
		ttl := binary.BigEndian.Uint32(response[offset+2:])
		fmt.Printf("  TTL: %d\n", ttl)
		offset += 6

		// Read the length of the data.
		if offset+2 > len(response) {
			fmt.Println("Invalid DNS response: not enough data to read data length")
			return
		}
		offset += 2
	}
}
