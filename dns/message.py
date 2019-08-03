
import random
import struct


class InvalidMessageException(Exception):
	pass


HEADER_FLAGS_QUERY			= 0x0000
HEADER_FLAGS_REPLY			= 0x0001

HEADER_FLAGS_OPCODE_QUERY	= 0x0000
HEADER_FLAGS_OPCODE_IQUERY	= 0x0000 #@@
HEADER_FLAGS_OPCODE_STATUS	= 0x0000 #@@

HEADER_FLAGS_AUTHORITATIVE_ANSWER	= 0x0020
HEADER_FLAGS_TRUNCATED				= 0x0040
HEADER_FLAGS_RECURSION_DESIRED		= 0x0080
HEADER_FLAGS_RECURSION_AVAILABLE	= 0x0100

HEADER_FLAGS_RESPONSE_SUCCESS			= 0x0000
HEADER_FLAGS_RESPONSE_FORMAT_ERROR		= 0x1000
HEADER_FLAGS_RESPONSE_SERVER_ERROR		= 0x2000
HEADER_FLAGS_RESPONSE_NAME_ERROR		= 0x3000
HEADER_FLAGS_RESPONSE_NOT_IMPLEMENTED	= 0x4000
HEADER_FLAGS_RESPONSE_REFUSED			= 0x5000

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


	def pack(self):
		"""
		Pack this message into a single bytestring, suitable
		for transmission.  Returns bytestring.  No side effects
		"""
		fields = [ self.header ] + self.question
		bytes = b"".join(field.pack() for field in fields)
		assert(len(bytes) <= MESSAGE_MAX_SIZE)
		return(bytes)


	@staticmethod
	def unpack(bytes):
		assert(bytes)
		(self.header, remainder) = Header.unpack(bytes)
		return



QUESTION_TYPE_A		= 1		# Host address
QUESTION_TYPE_NS	= 2		# Name server

QUESTION_CLASS_IN	= 1		# Internet
QUESTION_CLASS_CS	= 2		# CSNET, obsolete
QUESTION_CLASS_CH	= 3		# Chaos
QUESTION_CLASS_HS	= 4		# Hesiod
QUESTION_CLASS_ANY	= 255	# Any


class Question(object):
	"""
	Question section of a DNS query
	"""

	def __init__(self, name="", type=QUESTION_TYPE_A, qclass=QUESTION_CLASS_IN):
		self.name = name
		self.type = type
		self.qclass = qclass
		return

	def pack_label(self, label):
		"""
		Pack a single DNS label into a bytesstring suitable for a
		Question or RR.	 Returns the packed label string.
		No side effects.

		pack_label("www") => b"\03www"
		pack_label("google") => b"\06google"
		pack_label("") => b"\00"
		"""
		return struct.pack("!%dp" % (len(label) + 1), bytes(label, "ascii"))

	def pack(self):
		"""
		Packs a Question into a bytestring suitable for a
		a DNS message.	Returns the packed Question string.
		No side effects
		"""
		null_label = ""
		labels = self.name.split(".") + [ null_label ]
		packed_labels = [ self.pack_label(label) for label in labels ]
		qname = b"".join(packed_labels)
		return struct.pack("!%dsHH" % len(qname), qname, self.type, self.qclass)

	@staticmethod
	def unpack_label(bytes):
		try:
			(label,) = struct.unpack("!%dp" % len(bytes), bytes)
			remainder = bytes[ 1 + len(label): ] # length byte + label bytes
			return(label.decode("utf8"), remainder)
		except:
			raise InvalidMessageException("Invalid label")
