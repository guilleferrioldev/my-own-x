package main

import "encoding/binary"

// Structure representing a DNS header.
type DNSHeader struct {
	PackedIdentifier       uint16
	QuestionCount          uint16
	AnswerRecordCount      uint16
	AuthorityRecordCount   uint16
	AdditionalRecordCount  uint16
	QueryResponseIndicator bool
	OperationCode          byte
	AuthoritativeAnswer    bool
	Truncation             bool
	RecursionDesired       bool
	RecursionAvailable     bool
	CheckingDisabled       bool
	AuthedData             bool
	Reserved               bool
	ResponseCode           byte
}

// Encode a DNS header into binary format.
func decodeDNSHeader(receivedData []byte) (header DNSHeader) {
	header.PackedIdentifier = binary.BigEndian.Uint16(receivedData)
	header.QuestionCount = binary.BigEndian.Uint16(receivedData[4:])
	header.AnswerRecordCount = binary.BigEndian.Uint16(receivedData[6:])
	header.AuthorityRecordCount = binary.BigEndian.Uint16(receivedData[8:])
	header.AdditionalRecordCount = binary.BigEndian.Uint16(receivedData[10:])

	flags := binary.BigEndian.Uint16(receivedData[2:])
	a := byte(flags >> 8)
	b := byte(flags & 0xFF)

	header.RecursionDesired = (a & 0x01) != 0
	header.Truncation = (a & 0x02) != 0
	header.AuthoritativeAnswer = (a & 0x04) != 0
	header.OperationCode = (a >> 3) & 0x0F
	header.QueryResponseIndicator = (a & 0x80) != 0

	header.ResponseCode = b & 0x0F
	header.CheckingDisabled = (b & 0x10) != 0
	header.AuthedData = (b & 0x20) != 0
	header.Reserved = (b & 0x40) != 0
	header.RecursionAvailable = (b & 0x80) != 0

	return
}

func encodeDNSHeader(header DNSHeader) (response []byte) {
	flags := uint16(0)
	if header.RecursionDesired {
		flags |= 0x0100
	}
	if header.Truncation {
		flags |= 0x0200
	}
	if header.AuthoritativeAnswer {
		flags |= 0x0400
	}
	if header.QueryResponseIndicator {
		flags |= 0x8000
	}
	flags |= uint16(header.OperationCode) << 11
	flags |= uint16(header.ResponseCode)

	if header.CheckingDisabled {
		flags |= 0x10
	}
	if header.AuthedData {
		flags |= 0x20
	}
	if header.Reserved {
		flags |= 0x40
	}
	if header.RecursionAvailable {
		flags |= 0x80
	}

	// Serialize the header fields.
	response = binary.BigEndian.AppendUint16(response, header.PackedIdentifier)
	response = binary.BigEndian.AppendUint16(response, flags)
	response = binary.BigEndian.AppendUint16(response, header.QuestionCount)
	response = binary.BigEndian.AppendUint16(response, header.AnswerRecordCount)
	response = binary.BigEndian.AppendUint16(response, header.AuthorityRecordCount)
	response = binary.BigEndian.AppendUint16(response, header.AdditionalRecordCount)

	return
}
