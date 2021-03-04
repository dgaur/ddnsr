//
// DNS protocol + structures defined by RFC 1035 et al
//

package main

import (
	"bytes"
	"encoding/binary"
)


//
// Message header
//
type Header struct {
	Id					int16
	Flags				int16
	QuestionCount		int16
	AnswerCount			int16
	NameserverCount		int16
	AdditionalCount		int16
}
const HeaderSize = 12 // 6 fields, 16b each

//
// Pack header struct into a sequence of bytes.  No side effects.
//
func (header Header) pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, header)
	return buffer.Bytes(), nil
}

//
// Unpack a sequence of bytes into header struct.  No side effects.
//
func UnpackHeader(rawBytes []byte) (Header, error) {
	header := Header{}
	reader := bytes.NewReader(rawBytes)
	err := binary.Read(reader, binary.BigEndian, &header)
	return header, err
}


//
// Label manipulation
//
func packLabel(label string) []byte {
	buffer := new(bytes.Buffer)
	buffer.WriteByte(byte(len(label)))
	buffer.WriteString(label)
	return buffer.Bytes()
}

func unpackLabel(rawBytes []byte) string {
	length := int(rawBytes[0])
	return string(rawBytes[1:length+1])
}


//
// Question section
//
type Question struct {
	Name	[]byte
	Type	uint16
	Class	uint16
}

func (question Question) pack() ([]byte, error) {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, question)
	return buffer.Bytes(), nil
}


func resolve(host string) {
}
