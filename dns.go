//
// DNS protocol + structures defined by RFC 1035, 3596, et al
//

package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"
)


//
// DNS message header
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

const MessageHeaderFlagResponse				= 0x8000
const MessageHeaderFlagAuthoritative		= 0x0400
const MessageHeaderFlagTruncation			= 0x0200
const MessageHeaderFlagRecursionDesired		= 0x0100
const MessageHeaderFlagRecursionAvailable	= 0x0080
const MessageHeaderFlagResponseCodeMask		= 0x000F

func (header MessageHeader) String() string {
	// Expand flag fields into human-friendly codes
	var flags []string
	if (header.Flags & MessageHeaderFlagResponse != 0) {
		flags = append(flags, "QR")
	}
	if (header.Flags & MessageHeaderFlagAuthoritative != 0) {
		flags = append(flags, "AA")
	}
	if (header.Flags & MessageHeaderFlagTruncation != 0) {
		flags = append(flags, "TC")
	}
	if (header.Flags & MessageHeaderFlagRecursionDesired != 0) {
		flags = append(flags, "RD")
	}
	if (header.Flags & MessageHeaderFlagRecursionAvailable != 0) {
		flags = append(flags, "RA")
	}
	if (header.Flags & MessageHeaderFlagResponseCodeMask != 0) {
		flags = append(flags, "RCODE-ERROR")
	}

	return fmt.Sprintf(
    "flags %#04x (%s), QD %d, AN %d, NS %d, AR %d",
	header.Flags,
	strings.Join(flags, " "),
	header.QuestionCount,
	header.AnswerCount,
	header.NameserverCount,
	header.AdditionalCount)
}

func packMessageHeader(header MessageHeader) []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, header)
	return buffer.Bytes()
}

func unpackMessageHeader(rawBytes []byte, offset int) (MessageHeader, int, error) {
	header := MessageHeader{}
	reader := bytes.NewReader(rawBytes)
	err := binary.Read(reader, binary.BigEndian, &header)
	return header, MessageHeaderSize, err
}


//
// Composite domain-name manipulation
//
const DomainNameMaxLength	= 255
const LabelMaxLength		= 63

func packName(name string) []byte {
	// Break the domain name into individual labels
	labels := append(strings.Split(name, "."), "")

	// Pack the individual labels into a continuous byte sequence
	buffer := new(bytes.Buffer)
	for _, label := range labels {
		buffer.WriteByte(byte(len(label)))
		buffer.WriteString(label)
	}

	// Pack all of the labels into a single list of bytes
	return buffer.Bytes()
}

func unpackName(rawBytes []byte, offset int) (string, int) {
	compressed	:= false
	labels		:= []string{}
	length		:= 0

	for {
		labelLength := int(rawBytes[offset:][0])
		if (labelLength > LabelMaxLength) {
			// This is a compressed label.  The pointer consumes 2 bytes,
			// but no more bytes at this offset
			if (!compressed) {
				length += 2 // 2 offset bytes
			}
			compressed = true

			// Jump to the new offset and continue unpacking from there
			offset = int(binary.BigEndian.Uint16(rawBytes[offset:offset+2]))-0xC000
			continue
		} else if (labelLength == 0) {
			// Zero-length label.  This is the end of the domain name.
			if (!compressed) {
				length++ // Account for the trailing zero-byte
			}
			break
		} else {
			// Otherwise, this is a normal, inline label.  Not compressed.  Just
			// read the label directly
			label := string(rawBytes[offset+1:offset+labelLength+1])
			labels = append(labels, label)

			// Continue reading the next label
			offset += labelLength+1
			if (!compressed) {
				length += (labelLength + 1) // Include the length byte
			}
		}
	}

	// Assemble the individual labels into a full, dotted DNS name
	return strings.Join(labels, "."), length
}


//
// Common Question/Record types + constants.  Most of these are valid for
// both Question + ResourceRecord types
//
const RecordTypeA		= 1 
const RecordTypeNS		= 2
const RecordTypeCNAME	= 5
const RecordTypeSOA		= 6
const RecordTypePTR		= 12
const RecordTypeMX		= 15
const RecordTypeTXT		= 16
const RecordTypeAAAA	= 28
const RecordTypeALL		= 255

const RecordClassIN		= 1

var RecordTypeMapToType = map[string]uint16{
		"A":		RecordTypeA,
		"NS":		RecordTypeNS,
		"CNAME":	RecordTypeCNAME,
		"SOA":		RecordTypeSOA,
		"PTR":		RecordTypePTR,
		"MX":		RecordTypeMX,
		"TXT":		RecordTypeTXT,
		"AAAA":		RecordTypeAAAA,
		"ALL":		RecordTypeALL,
	}
var RecordTypeMapToString = map[uint16]string{}

func init() {
	for k, v := range RecordTypeMapToType {
		RecordTypeMapToString[v] = k
	}
}

//
// Question section
//
type Question struct {
	Name	string
	Type	uint16
	Class	uint16
}

func (question Question) String() string {
	var qtype string = RecordTypeMapToString[question.Type]
	if (qtype == "") {
		qtype = fmt.Sprintf("%d", int(question.Type))
	}

	return fmt.Sprintf("%s (%s)", question.Name, qtype)
}

func packQuestion(question Question) []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, packName(question.Name))
	binary.Write(buffer, binary.BigEndian, question.Type)
	binary.Write(buffer, binary.BigEndian, question.Class)
	return buffer.Bytes()
}

func unpackQuestion(rawBytes []byte, offset int) (Question, int, error) {
	var err error	= nil
	var length int	= 0
	var question	= Question{}

	// Parse the initial Name string, variable-length
	question.Name, length = unpackName(rawBytes, offset)

	// Parse the fixed fields after the Name string
	reader := bytes.NewReader(rawBytes[offset+length:])
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

//@@@this is wrong: each of these RR types really should be a subclass/extension
// of the main ResourceRecord struct.  For simplicity, just pack a few extra
// fields here instead instead of complicating the surrounding code.
type DecodedResourceRecord struct {
	CNAME		string
	MXExchange	string
	PTR			string
	NS			string
	SOAMNAME	string
	SOARNAME	string
}

type ResourceRecord struct {
	Name		string
	Type		uint16
	Class		uint16
	TTL			int32
	RDLength	uint16
	RData		[]byte

	Decoded		DecodedResourceRecord //@should really be subclassed
}


func (rr ResourceRecord) String() string {
	var rdata string
	var rtype string = RecordTypeMapToString[rr.Type]
	if (rtype == "") {
		rtype = fmt.Sprintf("%d", int(rr.Type))
	}

	// Where possible, decode the actual record payload
	switch (rr.Type) {
		case RecordTypeA:
			rdata = fmt.Sprintf("%d.%d.%d.%d",
				rr.RData[0], rr.RData[1], rr.RData[3], rr.RData[3])
		case RecordTypeCNAME:
			rdata = rr.Decoded.CNAME
		case RecordTypeMX:
			rdata = rr.Decoded.MXExchange
		case RecordTypeNS:
			rdata = rr.Decoded.NS
		case RecordTypePTR:
			rdata = rr.Decoded.PTR
		case RecordTypeSOA:
			rdata = fmt.Sprintf("mname %s, rname %s",
				rr.Decoded.SOAMNAME, rr.Decoded.SOARNAME)
		case RecordTypeTXT:
			rdata = string(rr.RData)
		default:
			rdata = fmt.Sprintf("rdata (%d bytes) % x", rr.RDLength, rr.RData)
	}

	return fmt.Sprintf("%s (%s), TTL %d: %s",
		rr.Name, rtype, rr.TTL, rdata)
}

func packResourceRecord(rr ResourceRecord) []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, packName(rr.Name))
	binary.Write(buffer, binary.BigEndian, rr.Type)
	binary.Write(buffer, binary.BigEndian, rr.Class)
	binary.Write(buffer, binary.BigEndian, rr.TTL)
	binary.Write(buffer, binary.BigEndian, rr.RDLength)
	binary.Write(buffer, binary.BigEndian, rr.RData)
	return buffer.Bytes()
}

func unpackResourceRecord(rawBytes []byte, offset int) (ResourceRecord, int, error) {
	var err error	= nil
	var length int	= 0
	var rr			= ResourceRecord{}

	// Parse the initial Name string, variable-length
	rr.Name, length = unpackName(rawBytes, offset)

	// RR type
	reader := bytes.NewReader(rawBytes[offset+length:])
	err = binary.Read(reader, binary.BigEndian, &rr.Type)
	if (err != nil) {
		fmt.Println("Unable to parse RR type: ", err)
		return ResourceRecord{}, 0, err
	}
	length += int(reflect.TypeOf(rr.Type).Size())

	// RR class
	err = binary.Read(reader, binary.BigEndian, &rr.Class)
	if (err != nil) {
		fmt.Println("Unable to parse RR class: ", err)
		return ResourceRecord{}, 0, err
	}
	length += int(reflect.TypeOf(rr.Class).Size())

	// RR TTL
	err = binary.Read(reader, binary.BigEndian, &rr.TTL)
	if (err != nil) {
		fmt.Println("Unable to parse RR TTL: ", err)
		return ResourceRecord{}, 0, err
	}
	length += int(reflect.TypeOf(rr.TTL).Size())

	// RR RDLength (payload length)
	err = binary.Read(reader, binary.BigEndian, &rr.RDLength)
	if (err != nil) {
		fmt.Println("Unable to parse RR RDLENGTH: ", err)
		return ResourceRecord{}, 0, err
	}
	length += int(reflect.TypeOf(rr.RDLength).Size())

	// RR RDLength (payload), variable-length
	rr.RData = make([]byte, rr.RDLength)
	err = binary.Read(reader, binary.BigEndian, &rr.RData)
	if (err != nil) {
		fmt.Println("Unable to parse RR RDATA: ", err)
		return ResourceRecord{}, 0, err
	}

	// Where possible, unpack some of the unique fields here when the RData is
	// readily available
	//@this is incomplete: missing fields, and would be better as subclasses
	switch (rr.Type) {
		case RecordTypeCNAME:
			rr.Decoded.CNAME, _ = unpackName(rawBytes, offset + length)
		case RecordTypeMX:
			rr.Decoded.MXExchange, _ = unpackName(rawBytes, offset + length + 2)
		case RecordTypeNS:
			rr.Decoded.NS, _    = unpackName(rawBytes, offset + length)
		case RecordTypePTR:
			rr.Decoded.PTR, _   = unpackName(rawBytes, offset + length)
		case RecordTypeSOA:
			var mlen int
			rr.Decoded.SOAMNAME, mlen = unpackName(rawBytes, offset + length)
			rr.Decoded.SOARNAME, _    = unpackName(rawBytes, offset + length + mlen)
	}

	// Include the payload bytes in the total, regardless of whether they
	// were unpacked or not
	length += int(rr.RDLength)

	return rr, length, err
}

//
// Composite DNS request/response messages
//
type Message struct {
	Header			MessageHeader
	Questions		[]Question
	Answers			[]ResourceRecord
	Nameservers		[]ResourceRecord
	AdditionalRR	[]ResourceRecord
}

func (message *Message) addQuestion(question Question) {
	// Add a new Question to a Request message, mostly useful for coordinating
	// changes to both the header and payload
	message.Header.QuestionCount++
	message.Questions = append(message.Questions, question)
}

func (message Message) String() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "H:  %s\n", message.Header)
	for _, q := range message.Questions {
		fmt.Fprintf(&builder, "Q:  %s\n", q)
	}
	for _, a := range message.Answers {
		fmt.Fprintf(&builder, "A:  %s\n", a)
	}
	for _, ns := range message.Nameservers {
		fmt.Fprintf(&builder, "NS: %s\n", ns)
	}
	for _, rr := range message.AdditionalRR {
		fmt.Fprintf(&builder, "RR: %s\n", rr)
	}
	return builder.String()
}

func (reply Message) validate(request Message) error {
	// Expect the request/response ids to match
	if (reply.Header.Id != request.Header.Id) {
		return(errors.New("Header id mismatch"))
	}

	// Expect a response message
	if (reply.Header.Flags & MessageHeaderFlagResponse == 0) {
		return(errors.New("Expected DNS response"))
	}

	// Abort on a truncated message, for simplicity
	if (reply.Header.Flags & MessageHeaderFlagTruncation != 0) {
		return(errors.New("DNS response truncated"))
	}

	// Here, the reply itself appears to be valid.  It may contain an
	// error message or unexpected RR, etc, but the message itself appears
	// to be well-formed, etc
	return(nil)
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
	message.Header, length, err = unpackMessageHeader(rawBytes, 0)
	if (err != nil) {
		fmt.Println("Unable to unpack header: ", err)
		return Message{}, 0, err
	}

	// Parse the Questions, if any
	for q := 0; q < int(message.Header.QuestionCount); q++ {
		question, questionLength, err := unpackQuestion(rawBytes, length)
		if (err != nil) {
			fmt.Println("Unable to unpack question: ", err)
			return Message{}, 0, err
		}
		message.Questions = append(message.Questions, question)

		// Continue parsing the next Question, if any
		length += questionLength
	}

	// Helper function for unmarshalling the different RR sections.  Just
	// collect all of the RR's embedded in an individual section
	unpackRRSection := func(count int) ([]ResourceRecord, error) {
		var records = []ResourceRecord{}

		for i := 0; i < count; i++ {
			// Attempt to parse the next RR in this section
			rr, rrLength, err := unpackResourceRecord(rawBytes, length)
			if (err != nil) {
				fmt.Println("Unable to unpack RR: ", err)
				return records, err
			}

			// Collect all of the RR's in this section
			records = append(records, rr)

			// Resume parsing at the next RR, if any
			length += rrLength
		}

		return records, nil
	}

	// Parse the Answers, if any
	message.Answers, err = unpackRRSection(int(message.Header.AnswerCount))
	if (err != nil) {
		return Message{}, 0, err
	}

	// Parse the Nameservers, if any
	message.Nameservers, err = unpackRRSection(int(message.Header.NameserverCount))
	if (err != nil) {
		return Message{}, 0, err
	}

	// Parse the related RR's, if any
	message.AdditionalRR, err = unpackRRSection(int(message.Header.AdditionalCount))
	if (err != nil) {
		return Message{}, 0, err
	}

	return message, length, nil
}

func dumpBytes(header string, rawBytes []byte) {
	fmt.Printf("%s: % x\n", header, rawBytes)
}


//
// Main resolver logic
//
const DnsPort = 53

func resolve(config ClientConfig, host string) error {
	// Locate the upstream DNS resolver
	upstreamPort := net.UDPAddr{
		IP:		net.ParseIP(config.server),
		Port:	DnsPort,
		Zone:	"",
	}
	upstream, err := net.DialUDP("udp", nil, &upstreamPort)
	if (err != nil) {
		fmt.Println("Unable to reach upstream DNS server: ", err)
		return err
	}
	defer upstream.Close()

	// Initialize the primitive DNS question for the upstream server
	question := Question{
		host,
		RecordTypeMapToType[config.rtype],
		RecordClassIN,
	}

	// Create the initial DNS request
	request := Message{}
	request.Header.Id = 0xFEFF
	if (config.recursive) {
		request.Header.Flags |= MessageHeaderFlagRecursionDesired
	}
	request.addQuestion(question)
	requestBytes := packMessage(request)
	if (config.raw) {
		dumpBytes("Raw request bytes", requestBytes)
	}

	// Send the actual DNS request
	_, err = upstream.Write(requestBytes)
	if (err != nil) {
		fmt.Println("Unable to send DNS request: ", err)
		return err
	}

	// Wait for a reply, if any
	replyBytes := make([]byte, 1024)
	timeout := time.Duration(config.timeout) * time.Second
	upstream.SetReadDeadline(time.Now().Add(timeout))
	length, err := upstream.Read(replyBytes)
	if err != nil {
		fmt.Println("Unable to read DNS response: ", err)
		return err
	}
	if (config.raw) {
		dumpBytes("Raw reply bytes", replyBytes[:length])
	}

	// Parse + validate the reply
	reply, _, err := unpackMessage(replyBytes)
	if (err != nil) {
		fmt.Println("Unable to parse DNS response: ", err)
		return err
	}
	err = reply.validate(request)
	if (err != nil) {
		fmt.Println("Invalid DNS response: ", err)
		return err
	}
	fmt.Println(reply)

	return err
}
