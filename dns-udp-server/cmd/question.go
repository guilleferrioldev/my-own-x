package main

import "encoding/binary"

// Structure representing a DNS question.
type DNSQuestion struct {
	Name  []string
	Type  uint16
	Class uint16
}

// Decode a DNS question from binary data
func decodeQuestion(data []byte, offset int) (question DNSQuestion, size int) {
	i := offset
	for i < len(data) {
		length := int(data[i])
		i++
		if length == 0 {
			break
		}
		if length <= 63 {
			// Add the label to the question name.
			question.Name = append(question.Name, string(data[i:i+length]))
			i += length
		} else {
			// Handle DNS name compression.
			compressedOffset := int(binary.BigEndian.Uint16(data[i-1:]) & 0b0011111111111111)
			j := compressedOffset
			for j < len(data) {
				length := int(data[j])
				j++
				if length == 0 {
					break
				}
				temp := string(data[j : j+length])
				question.Name = append(question.Name, temp)
				j += length
			}
			i++
			break
		}
	}
	question.Type = binary.BigEndian.Uint16(data[i:])
	i += 2
	question.Class = binary.BigEndian.Uint16(data[i:])
	i += 2
	size = i - offset
	return
}

// Encode a DNS question into binary format.
func encodeQuestion(question DNSQuestion) (response []byte) {
	for _, name := range question.Name {
		length := len(name)
		response = append(response, byte(length))
		response = append(response, []byte(name)...)
	}
	response = append(response, byte(0))
	response = binary.BigEndian.AppendUint16(response, question.Type)
	response = binary.BigEndian.AppendUint16(response, question.Class)
	return
}
