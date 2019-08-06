
import dns.message
import socket

DNS_PORT = 53


class Resolver(object):
	"""
	DNS resolver
	"""
	
	def __init__(self):
		self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
		return
		
	def query(self, name, server):
		request = dns.message.Message(name)
		destination = (server, DNS_PORT)

		# Send the initial query
		self.sock.sendto(request.pack(), destination)

		# Wait for the response
		self.sock.settimeout(2.0)
		(bytes, source) = self.sock.recvfrom(dns.message.MESSAGE_MAX_SIZE)
		print(bytes)
		response = dns.message.Message.unpack(bytes)
		assert(request.header.id == response.header.id)
		print(response)
		
		return

