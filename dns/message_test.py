
import message
import unittest

class MessageTest(unittest.TestCase):
	def test_header_packing(self):
		h = message.Header(question_count=3)
		bytes = h.pack()

		# Expect 6 half-words
		assert(len(bytes) == 6*2)

		# Ignore id field.	Expect RECURSION_DESIRED + 3 questions
		assert(bytes[2:] == b"\x00\x80\x00\x03\x00\x00\x00\x00\x00\x00")
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
		q = message.Question()
		assert(q.pack_label("www") == b"\03www")
		assert(q.pack_label("") == b"\00")
		return


	def test_message_packing(self):
		m1 = message.Message(name="www.google.com")
		assert(m1.header.question_count == 1)
		assert(len(m1.pack()) < message.MESSAGE_MAX_SIZE)

		m2 = message.Message(name=["www.google.com", "www.amazon.com"])
		assert(m2.header.question_count == 2)
		assert(len(m2.pack()) < message.MESSAGE_MAX_SIZE)
		return


	def test_question_packing(self):
		q = message.Question("www.google.com")
		assert(q.pack() == b"\03www\06google\03com\00\00\01\00\01")


if __name__ == "__main__":
	unittest.main()

