GoServer package

This folder is a ready-to-run local website package. It includes:

- `server.json`
- `web/templates/index.html`
- `web/static/styles.css`

How to start:

Windows:
server.exe

Linux or macOS:
./server

How to open it:

- Open `http://localhost:8081/` in your browser.
- To check that the server is running, use `navigate to http://localhost:8081/health`.

Edit your page:

- Change the HTML in `web/templates/index.html`.
- Change the CSS in `web/static/styles.css`.

How to stop:

- Press `Ctrl+C` in the terminal where the server is running.

Troubleshooting:

- Port already in use: stop the other program or change `server_address` in `server.json`.
- Browser cannot connect: check that the server is still running and the port is `8081`.
- Page changes are not visible: save the files and restart the server.
- Config paths are wrong: make sure `template_dir` is `web/templates` and `static_dir` is `web/static`.

