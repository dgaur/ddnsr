
import message
import unittest

class MessageTest(unittest.TestCase):
	def test_header_packing(self):
		h = message.Header(question_count=3)
		bytes = h.pack()

		# Expect 6 half-words
		assert(len(bytes) == 6*2)

		# Ignore id field.	Expect RECURSION_DESIRED + 3 questions
		assert(bytes[2:] == b"\x01\x00\x00\x03\x00\x00\x00\x00\x00\x00")
		return


	def test_header_roundtrip(self):
		h1 = message.Header(question_count=5)
		bytes = h1.pack()
		(h2, remainder) = message.Header.unpack(bytes)
		assert(h1.id == h2.id)
		assert(h1.question_count == h2.question_count)
		assert(len(remainder) == 0)
		return


	def test_label_packing(self):
		n = message.DomainName()
		assert(n.pack_label("www") == b"\03www")
		assert(n.pack_label("") == b"\00")
		return


	def test_label_unpacking(self):
		n = message.DomainName()
		(label, remainder) = n.unpack_label(b"\03www\06google")
		assert(label == "www")
		assert(len(remainder) == 1+6)
		return


	def test_message_packing(self):
		m1 = message.Message(name="www.google.com")
		assert(m1.header.question_count == 1)
		assert(len(m1.pack()) < message.MESSAGE_MAX_SIZE)

		m2 = message.Message(name=["www.google.com", "www.amazon.com"])
		assert(m2.header.question_count == 2)
		assert(len(m2.pack()) < message.MESSAGE_MAX_SIZE)
		return


	def test_name_packing(self):
		n = message.DomainName("www.amazon.com")
		assert(n.pack() == b"\03www\06amazon\03com\0")
		return


	def test_name_unpacking(self):
		(n, remainder) = message.DomainName.unpack(b"\03www\06amazon\03com\0\xFF\xFF")
		assert(str(n) == "www.amazon.com")
		assert(remainder == b"\xFF\xFF")
		return

	def test_question_packing(self):
		q = message.Question("www.google.com")
		assert(q.pack() == b"\03www\06google\03com\00\00\01\00\01")


if __name__ == "__main__":
	unittest.main()

