//
// DNS protocol + structures defined by RFC 1035 et al
//

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"
)

func swap16(data16 uint16) uint16 {
	return ((data16 & 0xFF00) >> 8) | ((data16 & 0xFF) << 8)
}


//
// Message header
//
type MessageHeader struct {
	Id					uint16
	Flags				uint16
	QuestionCount		uint16
	AnswerCount			uint16
	NameserverCount		uint16
	AdditionalCount		uint16
}
const MessageHeaderSize = 12 // 6 fields, 16b each

const MessageHeaderFlagRecursionDesired = 0x0100

func (header MessageHeader) dump() {
	fmt.Printf("Header:\n")
	fmt.Printf("   Id:           %#04x\n", header.Id)
	fmt.Printf("   Flags:        %#04x\n", header.Flags)
	fmt.Printf("   Questions:    %d\n", header.QuestionCount)
	fmt.Printf("   Answers:      %d\n", header.AnswerCount)
	fmt.Printf("   Nameservers:  %d\n", header.NameserverCount)
	fmt.Printf("   AdditionalRR: %d\n", header.AdditionalCount)
	return
}

func packMessageHeader(header MessageHeader) []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, header)
	return buffer.Bytes()
}

func unpackMessageHeader(rawBytes []byte) (MessageHeader, int, error) {
	header := MessageHeader{}
	reader := bytes.NewReader(rawBytes)
	err := binary.Read(reader, binary.BigEndian, &header)
	return header, MessageHeaderSize, err
}


//
// Individual label manipulation
//
const LabelMaxLength = 63

func packLabel(label string) []byte {
	buffer := new(bytes.Buffer)
	buffer.WriteByte(byte(len(label)))
	buffer.WriteString(label)
	return buffer.Bytes()
}

func unpackLabel(label []byte) (string, int) {
	length := int(label[0])
	return string(label[1:length+1]), length
}


//
// Composite domain-name manipulation
//
const DomainNameMaxLength = 255

func packName(name string) []byte {
	// Break the domain name into individual labels
	tokens := append(strings.Split(name, "."), "")

	// Compute the individual labels
	labels := [][]byte{}
	for _, token := range tokens {
		labels = append(labels, packLabel(token))
	}

	// Pack all of the labels into a single list of bytes
	return bytes.Join(labels, []byte{})
}

func unpackName(domainName []byte) (string, int) {
	tokens := []string{}
	length := 0
	for {
		// Unpack the next label in the domain name
		token, tokenLength := unpackLabel(domainName[length:])

		// Collect the labels until the empty label
		if tokenLength > 0 {
			tokens = append(tokens, token)
			length += (tokenLength + 1) // Include the length byte
		} else {
			break
		}
	}

	return strings.Join(tokens, "."), length
}


//
// Question section
//
const QuestionTypeA		= 1 //@these are really RR types and classes
const QuestionClassIN	= 1

type Question struct {
	Name	string
	Type	uint16
	Class	uint16
}

func (question Question) dump() {
	fmt.Printf("Question:\n")
	fmt.Printf("   Name:         %s\n", question.Name)
	fmt.Printf("   Type:         %#02x\n", question.Type)
	fmt.Printf("   Class:        %#02x\n", question.Class)
}

func packQuestion(question Question) []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, packName(question.Name))
	binary.Write(buffer, binary.BigEndian, question.Type)
	binary.Write(buffer, binary.BigEndian, question.Class)
	return buffer.Bytes()
}

func unpackQuestion(rawBytes []byte) (Question, int, error) {
	var err error	= nil
	var length int	= 0
	var question	= Question{}

	// Parse the initial Name string, variable-length
	question.Name, length = unpackName(rawBytes)

	// Parse the fixed fields after the Name string
	reader := bytes.NewReader(rawBytes[length:])
	err = binary.Read(reader, binary.BigEndian, &question.Type)
	if (err != nil) {
		fmt.Println("Unable to parse question type: ", err)
		return Question{}, 0, err
	}
	length += int(reflect.TypeOf(question.Type).Size())
	
	err = binary.Read(reader, binary.BigEndian, &question.Class)
	if (err != nil) {
		fmt.Println("Unable to parse question class: ", err)
		return Question{}, 0, err
	}
	length += int(reflect.TypeOf(question.Class).Size())

	return question, length, err
}


//
// Resource records (RR)
//
type ResourceRecord struct {
}

func unpackResourceRecord(rawBytes []byte) (ResourceRecord, int, error) {
	var err error	= nil
	var length int	= 0
	var rr			= ResourceRecord{}
	
	return rr, length, err
}

//
// Composite DNS request/response
//
type Message struct {
	Header		MessageHeader
	Questions	[]Question
}

func (message *Message) addQuestion(question Question) {
	message.Header.QuestionCount++
	message.Questions = append(message.Questions, question)
}

func (message Message) dump() {
	message.Header.dump()
	var q uint16 = 0
	for (q < message.Header.QuestionCount) {
		message.Questions[q].dump()
		q++
	}
}

func packMessage(message Message) []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, packMessageHeader(message.Header))
	for _, question := range message.Questions {
		binary.Write(buffer, binary.BigEndian, packQuestion(question))
	}

	return buffer.Bytes()
}

func unpackMessage(rawBytes []byte) (Message, int, error) {
	var err error	= nil
	var length int	= 0
	var message		= Message{}

	// Message header is always present
	message.Header, length, err = unpackMessageHeader(rawBytes)
	if (err != nil) {
		fmt.Println("Unable to unpack header: ", err)
		return Message{}, 0, err
	}

	// Parse the Questions, if any
	var q uint16 = 0
	for (q < message.Header.QuestionCount) {
		question, questionLength, err := unpackQuestion(rawBytes[length:])
		if (err != nil) {
			fmt.Println("Unable to unpack question: ", err)
			return Message{}, 0, err
		}
		message.Questions = append(message.Questions, question)

		// Continue parsing the next Question, if any
		q++
		length += questionLength
	}

	return message, length, err
}

func dumpBytes(rawBytes []byte) {
	fmt.Printf("Raw bytes: ")
	for _, byte := range rawBytes {
		fmt.Printf("%02x ", byte)
	}
	fmt.Printf("\n")
}


//
// Main resolver logic
//
const DnsPort = 53

func resolve(host string) error {
	// Locate the upstream DNS resolver
	upstreamPort := net.UDPAddr{ net.ParseIP("8.8.8.8"), DnsPort, ""}
	upstream, err := net.DialUDP("udp", nil, &upstreamPort)
	if (err != nil) {
		fmt.Println("Unable to reach upstream DNS server: ", err)
		return err
	}
	defer upstream.Close()

	// Create the DNS request
	request := Message{}
	request.Header.Id = 0xFEFF
	request.Header.Flags |= MessageHeaderFlagRecursionDesired
	request.addQuestion( Question{ host, QuestionTypeA, QuestionClassIN } )
	requestBytes := packMessage(request)
	dumpBytes(requestBytes)

	// Send the actual DNS request
	_, err = upstream.Write(requestBytes)
	if (err != nil) {
		fmt.Println("Unable to send DNS request: ", err)
		return err
	}

	// Wait for a reply, if any
	replyBytes := make([]byte, 1024)
	upstream.SetReadDeadline(time.Now().Add(5 * time.Second))
	length, err := upstream.Read(replyBytes)
	if err != nil {
		fmt.Println("Unable to read DNS response: ", err)
		return err
	}
	dumpBytes(replyBytes[:length])

	reply, _, err := unpackMessage(replyBytes)
	if (err != nil) {
		fmt.Println("Unable to parse DNS response: ", err)
		return err
	}
	reply.dump()

	return err
}
