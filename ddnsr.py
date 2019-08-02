
import argparse
import dns.resolver
import sys

def parse_command_line():
	parser = argparse.ArgumentParser()
	args = parser.parse_args()
	return(args)


if __name__ == "__main__":
	status = 0
	try:
		config = parse_command_line()
		resolver = dns.resolver.Resolver()
		resolver.query("google.com", "8.8.8.8")
		
	except KeyboardInterrupt:
		pass
		
	except SystemExit:
		pass
		
	except:
		import traceback
		traceback.print_exc()
		status = 1
		
	sys.exit(status)
	
