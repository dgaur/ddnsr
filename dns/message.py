
import random
import struct



#
# INTERNAL UTILITY ROUTINES ###############################################
#

class InvalidMessageException(Exception):
	pass


def cat(*details):
	"""Convenience routine for concatenating multi-line fields"""
	return "\n".join(details)


def decode_all(mask, decoder):
	"""Safely decode a mask to its string equivalent"""
	result = [ v for (k,v) in decoder.items() if bool(k & mask) ]
	return ", ".join(result)


def decode_one(value, decoder):
	"""Safely decode a single value to its string equivalent"""
	try:
		return decoder[value]
	except:
		return "Unknown (%#x)" % value




#
# DOMAIN NAME/LABELS ######################################################
#

class DomainName(object):
	def __init__(self, name=""):
		self.name = name
		return

	def pack(self):
		"""
		Pack a DNS name into a byte-string suitable for a Question or RR.
		Returns the packed name.  No side effects.
		"""
		null_label = ""
		labels = self.name.split(".") + [ null_label ]
		packed_labels = [ self.pack_label(label) for label in labels ]
		qname = b"".join(packed_labels)
		return(qname)

	def pack_label(self, label):
		"""
		Pack a single DNS label into a byte-string suitable for a
		Question or RR.	 Returns the packed label string.
		No side effects.

		pack_label("www") => b"\03www"
		pack_label("google") => b"\06google"
		pack_label("") => b"\00"
		"""
		return struct.pack("!%dp" % (len(label) + 1), bytes(label, "ascii"))

	def __str__(self):
		return self.name

	@staticmethod
	def unpack(bytes):
		"""
		Unpack a single DNS name from a string of bytes.
		Returns (DomainName, remaining-bytes)
		"""
		remainder = bytes

		# Unpack the entire domain name, piece-by-piece
		labels = []
		while True:
			(label, remainder) = DomainName.unpack_label(remainder)
			if (len(label) > 0):
				labels.append(label)
			else:
				break
		assert(len(labels) > 0)

		# Compute the complete name, from the individual labels
		name = ".".join(labels)

		return (DomainName(name), remainder)

	@staticmethod
	def unpack_label(bytes):
		try:
			(label,) = struct.unpack("!%dp" % len(bytes), bytes)
			remainder = bytes[ 1 + len(label): ] # length byte + label bytes
			return(label.decode("utf8"), remainder)
		except:
			raise InvalidMessageException("Invalid label")



#
# DNS MESSAGE HEADER ######################################################
#


HEADER_FLAGS_QUERY			= 0x0000
HEADER_FLAGS_REPLY			= 0x8000

HEADER_FLAGS_OPCODE_QUERY	= 0x0000
HEADER_FLAGS_OPCODE_IQUERY	= 0x0800
HEADER_FLAGS_OPCODE_STATUS	= 0x1000

HEADER_FLAGS_AUTHORITATIVE_ANSWER	= 0x0400
HEADER_FLAGS_TRUNCATED				= 0x0200
HEADER_FLAGS_RECURSION_DESIRED		= 0x0100
HEADER_FLAGS_RECURSION_AVAILABLE	= 0x0080

HEADER_FLAGS_RESPONSE_SUCCESS			= 0x0000
HEADER_FLAGS_RESPONSE_FORMAT_ERROR		= 0x0001
HEADER_FLAGS_RESPONSE_SERVER_ERROR		= 0x0002
HEADER_FLAGS_RESPONSE_NAME_ERROR		= 0x0003
HEADER_FLAGS_RESPONSE_NOT_IMPLEMENTED	= 0x0004
HEADER_FLAGS_RESPONSE_REFUSED			= 0x0005

HEADER_FLAGS_DECODER = {
	HEADER_FLAGS_QUERY:					"MESSAGE_QUERY",
	HEADER_FLAGS_REPLY:					"MESSAGE_REPLY",
	HEADER_FLAGS_OPCODE_QUERY:			"OPCODE_QUERY",
	HEADER_FLAGS_OPCODE_IQUERY:			"OPCODE_IQUERY",
	HEADER_FLAGS_OPCODE_STATUS:			"OPCODE_STATUS",
	HEADER_FLAGS_AUTHORITATIVE_ANSWER:	"AUTHORITATIVE",
	HEADER_FLAGS_TRUNCATED:				"TRUNCATED",
	HEADER_FLAGS_RECURSION_DESIRED:		"RECURSION_DESIRED",
	HEADER_FLAGS_RECURSION_AVAILABLE:	"RECURSION_AVAILABLE",
	HEADER_FLAGS_RESPONSE_SUCCESS:		"RESPONSE_SUCCESS",
	HEADER_FLAGS_RESPONSE_FORMAT_ERROR:	"RESPONSE_FORMAT_ERROR",
	HEADER_FLAGS_RESPONSE_SERVER_ERROR:	"RESPONSE_SERVER_ERROR",
	HEADER_FLAGS_RESPONSE_NAME_ERROR:	"RESPONSE_NAME_ERROR",
	HEADER_FLAGS_RESPONSE_NOT_IMPLEMENTED: "RESPONSE_NOT_IMPLEMENTED",
	HEADER_FLAGS_RESPONSE_REFUSED:		"RESPONSE_REFUSED"
}

HEADER_SIZE = 6*2	# Six half-words

class Header(object):
	"""
	DNS message header
	"""
	def __init__(self, flags=HEADER_FLAGS_RECURSION_DESIRED, question_count=0):
		self.id					= random.randint(0, 65535)	# 16 bits
		self.flags				= flags
		self.question_count		= question_count # QDCOUNT
		self.answer_count		= 0			# ANCOUNT
		self.nameserver_count	= 0			# NSCOUNT
		self.additional_count	= 0			# ARCOUNT
		return

	def pack(self):
		bytes = struct.pack("!HHHHHH",
							self.id,
							self.flags,
							self.question_count,
							self.answer_count,
							self.nameserver_count,
							self.additional_count)
		return(bytes)

	def __str__(self):
		return cat("Header:",
			"  id:        %#02x" % self.id,
			"  flags:     %#02x (%s)" % (self.flags, decode_all(self.flags, HEADER_FLAGS_DECODER)),
			"  questions: %d" % self.question_count,
			"  answers:   %d" % self.answer_count)


	@staticmethod
	def unpack(bytes):
		try:
			h = Header()
			(h.id,
			 h.flags,
			 h.question_count,
			 h.answer_count,
			 h.nameserver_count,
			 h.additional_count) = struct.unpack("!HHHHHH", bytes[:HEADER_SIZE])
			return(h, bytes[HEADER_SIZE:])

		except:
			raise InvalidMessageException()



#
# ACTUAL DNS MESSAGE ######################################################
#

MESSAGE_MAX_SIZE = 512	# 512 bytes, section 4.2.1 of RFC 1035

class Message(object):
	"""
	A single DNS Message, either query or response
	"""

	def __init__(self, name=[]):
		if (isinstance(name, str)):
			# Single domain name
			self.question = [ Question(name) ]
		else:
			# Multiple domain names
			assert(isinstance(name, list))
			self.question = [ Question(n) for n in name ]

		self.header = Header(question_count=len(self.question))
		return

	def fields(self):
		"""List of message fields, in order of packet layout"""
		return [ self.header ] + self.question

	def pack(self):
		"""
		Pack this message into a single bytestring, suitable
		for transmission.  Returns bytestring.  No side effects
		"""
		bytes = b"".join(field.pack() for field in self.fields())
		assert(len(bytes) <= MESSAGE_MAX_SIZE)
		return(bytes)


	def __str__(self):
		return "\n".join(str(field) for field in self.fields())


	@staticmethod
	def unpack(bytes):
		assert(bytes)
		assert(len(bytes) <= MESSAGE_MAX_SIZE)

		m = Message()
		(m.header, remainder) = Header.unpack(bytes)
		for i in range(m.header.question_count):
			(question, remainder) = Question.unpack(remainder)
			m.question.append(question)

		return m



#
# "QUESTION" STRUCTURE WITHIN DNS MESSAGE #################################
#


QUESTION_TYPE_A		= 1		# Host address
QUESTION_TYPE_NS	= 2		# Name server

QUESTION_TYPE_DECODER = {
	QUESTION_TYPE_A:	"QTYPE_A",
	QUESTION_TYPE_NS:	"QTYPE_NS"
}


QUESTION_CLASS_IN	= 1		# Internet
QUESTION_CLASS_CS	= 2		# CSNET, obsolete
QUESTION_CLASS_CH	= 3		# Chaos
QUESTION_CLASS_HS	= 4		# Hesiod
QUESTION_CLASS_ANY	= 255	# Any

QUESTION_CLASS_DECODER = {
	QUESTION_CLASS_IN:	"QCLASS_IN",
	QUESTION_CLASS_CS:	"QCLASS_CS",
	QUESTION_CLASS_CH:	"QCLASS_CH",
	QUESTION_CLASS_HS:	"QCLASS_HS",
	QUESTION_CLASS_ANY:	"QCLASS_ANY"
}


class Question(object):
	"""
	Question section of a DNS query
	"""

	def __init__(self, name="", type=QUESTION_TYPE_A, qclass=QUESTION_CLASS_IN):
		self.name = DomainName(name)
		self.type = type
		self.qclass = qclass
		return

	def pack(self):
		"""
		Packs a Question into a bytestring suitable for a
		a DNS message.	Returns the packed Question string.
		No side effects
		"""
		qname = self.name.pack()
		return struct.pack("!%dsHH" % len(qname), qname, self.type, self.qclass)

	def __str__(self):
		return cat("Question:",
			"  label:     %s" % self.name,
			"  type:      %#02x (%s)" % (self.type, decode_one(self.type, QUESTION_TYPE_DECODER)),
			"  class:     %#02x (%s)" % (self.qclass, decode_one(self.qclass, QUESTION_CLASS_DECODER)))


	@staticmethod
	def unpack(bytes):
		"""
		Unpack a single question from a string of bytes.
		Returns (Question, remaining-bytes)
		"""
		try:
			q = Question()
			(q.name, remainder) = DomainName.unpack(bytes)
			(q.type, q.qclass) = struct.unpack("!HH", remainder[:4])
			return(q, remainder[4:])

		except InvalidMessageException:
			raise
		except:
			raise InvalidMessageException("Invalid question")


