Critical: Stability
	- explicit error handling before accessing response properties
	- Manual cookieJar updates following handshake process
Features: WAF Bypass
	- Create auto-refresh logic to trigger new handshake on 403 status codes
	- Integrate headless browser automation to handle JavaScript base challenges
	- Add cli flags for dynamic proxy configuration + TLS profile selection
Future: Discovery
	- Implement passive discovery to extract new endpoints from html response bodys
	- Develop intellegent wordlist generation based on site path structure
	- Create a report generator to export successfull findings to local files
	- add CLI version and integration.
