GoServer release package template

This folder is meant to be shipped with the server executable.

What to do:
1. Put the executable in this folder.
2. Start it from this folder.
3. Open http://localhost:8081/ in your browser.
4. Edit the files under web/templates and web/static if you want to change the page.

Start the server:

Windows:
server.exe

Linux or macOS:
./server

Check that it is running:
curl http://localhost:8081/health

Edit these files:
- web/templates/index.html
- web/static/styles.css

Stop the server:
- Press Ctrl+C in the terminal running the server.

Common problems:
- Port already in use: stop the other program or change server_address in server.json.
- Browser cannot connect: check that the server is still running and the port is 8081.
- Page does not update: save the file and restart the server.
- Config paths are wrong: check template_dir and static_dir in server.json.

This package is a simple local site template, not production hosting.
