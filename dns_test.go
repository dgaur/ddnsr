package main

import(
	"bytes"
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


func TestLabelPacking(t *testing.T) {
	label := "label"

	// Pack into raw bytes
	rawBytes := packLabel(label)
	expected := []byte{ 5, 'l', 'a', 'b', 'e', 'l'}
	if !bytes.Equal(rawBytes, expected) {
		t.Error("Unexpected packed label: ", rawBytes)
	}

	// Deliberately append some additional text to the packed bytes, to
	// ensure the unpacking logic only consumes the correct length
	invalid := []byte("INVALID")
	rawBytes = append(rawBytes, invalid...)

	// Unpack back to a text string
	unpackedLabel := unpackLabel(rawBytes)
	if unpackedLabel != label {
		t.Error("Unexpected unpacked label: ", unpackedLabel)
	}
}


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

func TestNullLabel(t *testing.T) {
	label := ""

	// Pack into raw bytes
	rawBytes := packLabel(label)
	expected := []byte{ 0 }
	if !bytes.Equal(rawBytes, expected) {
		t.Error("Unexpected non-zero label: ", rawBytes)
	}

	// Unpack back to a text string
	unpackedLabel := unpackLabel(rawBytes)
	if unpackedLabel != label {
		t.Error("Unexpected non-empty label: ", unpackedLabel)
	}
}
