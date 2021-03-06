package main

import(
	"bytes"
	"testing"
	)

//
// Validate individual label packing/unpacking
//
func TestLabelPacking(t *testing.T) {
	testCases := []struct{
		name			string
		unpackedLabel	string
		packedLabel		[]byte
	}{
		{ "a",      "a",      []byte{ 1, 'a' } },
		{ "amazon", "amazon", []byte{ 6, 'a', 'm', 'a', 'z', 'o', 'n' } },
		{ "null",   "",       []byte{ 0 } },
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			// Pack and verify the label
			packedLabel := packLabel(test.unpackedLabel)
			if !bytes.Equal(packedLabel, test.packedLabel) {
				t.Error("Unexpected packed name: ", packedLabel,
					test.packedLabel)
			}

			// Unpack the label and revalidate
			unpackedLabel := unpackLabel(packedLabel)
			if unpackedLabel != test.unpackedLabel {
				t.Error("Unexpected unpacked name: ", unpackedLabel,
					test.unpackedLabel)
			}
		})
	}
}


//
// Validate header packing + unpacking
//
func TestMessageHeaderPacking(t *testing.T) {
	// Pack a trivial header into bytes
	header1 := MessageHeader{0,1,2,3,4,5}
	rawBytes := packMessageHeader(header1)
	if len(rawBytes) != MessageHeaderSize {
		t.Error("Packed header size is ", len(rawBytes))
	}

	// Unpack the bytes back into a new header and compare to the original.
	// Expect the two headers to be identical
	header2, err := unpackMessageHeader(rawBytes)
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


//
// Validate full message packing
//
func TestMessagePacking(t *testing.T) {
	message := Message{}
	message.Header.Id = 0xFEFF
	message.Header.Flags |= MessageHeaderFlagRecursionDesired
	message.addQuestion( Question{ "a.com", 0, 0 } )
	rawBytes := packMessage(message)
	t.Log("packed message: ", rawBytes)
}


//
// Validate composite domain-name packing/unpacking
//
func TestNamePacking(t *testing.T) {
	testCases := []struct{
		name			string
		unpackedName	string
		packedName		[]byte
	}{
		{ "a",      "a.com",      []byte{ 1, 'a', 3, 'c', 'o', 'm', 0 } },
		{ "amazon", "amazon.com", []byte{ 6, 'a', 'm', 'a', 'z', 'o', 'n', 3, 'c', 'o', 'm', 0 } },
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			// Pack and verify the domain name
			packedName := packName(test.unpackedName)
			if !bytes.Equal(packedName, test.packedName) {
				t.Error("Unexpected packed name: ", packedName, test.packedName)
			}

			// Unpack the labels and revalidate
			unpackedName := unpackName(packedName)
			if unpackedName != test.unpackedName {
				t.Error("Unexpected unpacked name: ", unpackedName,
					test.unpackedName)
			}
		})
	}

}


func TestQuestionPacking(t *testing.T) {
	question := Question{ "a.com", 0, 0 }
	packedQuestion := packQuestion(question)
	if len(packedQuestion) != ((1+1+1+3+1) + 2 + 2) {
		t.Error("Unexpected packed question length: ", len(packedQuestion))
	}
}
