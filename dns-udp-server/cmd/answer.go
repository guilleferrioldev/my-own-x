package main

import "encoding/binary"

// Structure representing a DNS answer.
type DNSAnswer struct {
	Name  []string
	Type  uint16
	Class uint16
	TTL   uint32
	Data  []byte
}

func encodeAnswer(answer DNSAnswer) (response []byte) {
	// Encode the domain name (using the same method as encoding the question).
	for _, name := range answer.Name {
		length := len(name)
		response = append(response, byte(length))
		response = append(response, []byte(name)...)
	}
	response = append(response, byte(0)) // End of domain name

	// Encode Type, Class, TTL, Data length and Data.
	response = binary.BigEndian.AppendUint16(response, answer.Type)
	response = binary.BigEndian.AppendUint16(response, answer.Class)
	response = binary.BigEndian.AppendUint32(response, answer.TTL)               // Correctly encode TTL as 32-bit
	response = binary.BigEndian.AppendUint16(response, uint16(len(answer.Data))) // Data length
	response = append(response, answer.Data...)                                  // Append the IP address or data (8.8.8.8)

	return
}
