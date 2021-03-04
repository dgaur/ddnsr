package main

import(
	"testing"
	)

//
// Validate header packing + unpacking
//
func TestHeaderPacking(t *testing.T) {
	// Pack a trivial header into bytes
	header1 := Header{0,1,2,3,4,5}
	rawBytes, err := header1.pack()
	if err != nil {
		t.Error("Packing error: ", err)
	}
	if len(rawBytes) != HeaderSize {
		t.Error("Packed header size is ", len(rawBytes))
	}

	// Unpack the bytes back into a new header and compare to the original.
	// Expect the two headers to be identical
	header2, err := UnpackHeader(rawBytes)
	if err != nil {
		t.Error("Unpacking error: ", err)
	}
	if (header1.Id != header2.Id) {
		t.Error("Id mismatch")
	}
	if (header1.Flags != header2.Flags) {
		t.Error("Flags mismatch", header1.Flags, header2.Flags)
	}
	if (header1.QuestionCount != header2.QuestionCount) {
		t.Error("QuestionCount mismatch", header1.QuestionCount, header2.QuestionCount)
	}
	if (header1.AnswerCount != header2.AnswerCount) {
		t.Error("AnswerCount mismatch", header1.AnswerCount, header2.AnswerCount)
	}
}

