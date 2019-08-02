
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
		m = dns.message.Message(name)
		host = (server, DNS_PORT)

		# Send the initial query
		self.sock.sendto(m.pack(), host)

		# Wait for the response
		self.sock.settimeout(2.0)
		(bytes, address) = self.sock.recvfrom(dns.message.MESSAGE_MAX_SIZE)
		print(bytes)
		
		return

