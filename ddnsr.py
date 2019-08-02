
import argparse
import dns.resolver
import sys

def parse_command_line():
	parser = argparse.ArgumentParser(description="DNS Resolver")

	parser.add_argument("name", nargs="+",
		help="DNS names(s) to resolve")
	parser.add_argument("--server", "-s", default="8.8.8.8",
		help="Upstream DNS server/resolver")

	args = parser.parse_args()
	return(args)


if __name__ == "__main__":
	status = 0
	try:
		config = parse_command_line()
		resolver = dns.resolver.Resolver()
		resolver.query(config.name, config.server)
		
	except KeyboardInterrupt:
		pass
		
	except SystemExit:
		pass
		
	except:
		import traceback
		traceback.print_exc()
		status = 1
		
	sys.exit(status)
	
