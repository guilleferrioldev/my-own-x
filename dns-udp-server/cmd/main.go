package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func dial(resolverAddr string) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{
			Timeout: time.Millisecond * time.Duration(10000),
		}
		return d.DialContext(ctx, network, resolverAddr)
	}
}

func main() {
	var resolverAddr string
	flag.StringVar(&resolverAddr, "resolver", "", "set resolver address")
	flag.Parse()

	if !flag.Parsed() {
		flag.Usage()
		return
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}
	log.Println("Running DNS server at port 2053")

	var fwdResolver *net.Resolver
	if resolverAddr != "" {
		fmt.Println("connecting to ", resolverAddr)
		fwdResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				dialer := &net.Dialer{
					Timeout: time.Second * time.Duration(10),
				}
				return dialer.DialContext(ctx, network, resolverAddr)
			},
		}
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		receivedData := string(buf[:size])
		fmt.Printf("Received %d bytes from %s: %q\n", size, source, receivedData)

		if size < 12 {
			fmt.Println("Unexpected size:", size)
			break
		}

		receivedHeader := decodeDNSHeader([]byte(receivedData))

		responseHeader := receivedHeader
		responseHeader.QueryResponseIndicator = true
		if receivedHeader.OperationCode == 0 {
			responseHeader.ResponseCode = 0
		} else {
			responseHeader.ResponseCode = 4
		}

		offset := 12 // skipping header

		encodedQuestions := []byte{}
		encodedAnswers := []byte{}

		// log.Println("receivedHeader.QuestionCount", receivedHeader.QuestionCount)
		for i := 0; i < int(receivedHeader.QuestionCount); i++ {
			// log.Println("question #", i)
			receivedQuestion, questionLength := decodeQuestion([]byte(receivedData), offset)
			// log.Printf("%+v\n", receivedQuestion)

			answer := DNSAnswer{
				Name:  receivedQuestion.Name,
				Type:  receivedQuestion.Type,
				Class: receivedQuestion.Class,
				TTL:   60,
				Data:  []byte{8, 8, 8, 8},
			}
			// log.Printf("%+v\n", answer)

			encodedQuestion := encodeQuestion(receivedQuestion)
			encodedQuestions = append(encodedQuestions, encodedQuestion...)

			if fwdResolver != nil {
				host := strings.Join(receivedQuestion.Name, ".")

				log.Println("forward request for", host)
				ips, err := fwdResolver.LookupIP(context.Background(), "ip4", host)
				if err == nil {
					log.Println("Got", ips)
					for _, ip := range ips {
						answer.Data = ip.To4()
						encodedAnswer := encodeAnswer(answer)
						encodedAnswers = append(encodedAnswers, encodedAnswer...)
						responseHeader.AnswerRecordCount++
					}
				} else {
					log.Println(err)
					answer.Data = []byte{8, 8, 8, 8}
					encodedAnswer := encodeAnswer(answer)
					encodedAnswers = append(encodedAnswers, encodedAnswer...)
					responseHeader.AnswerRecordCount++
					continue
				}
			}

			offset += questionLength
		}

		encodedHeader := encodeDNSHeader(responseHeader)

		response := []byte{}
		response = append(response, encodedHeader...)
		response = append(response, encodedQuestions...)
		response = append(response, encodedAnswers...)

		_, err = udpConn.WriteToUDP(response, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
