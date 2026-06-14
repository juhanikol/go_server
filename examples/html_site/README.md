# HTML Site Example

Serve a simple HTML/CSS page locally.

## Run

From the repository root:

```bash
go run ./examples/html_site
```

Then open:

```text
http://localhost:8083
```

## Edit The Page

Replace these files with your own page:

- `site/index.html`
- `site/styles.css`

Restart the server after changing files.

## Stop The Server

Press `Ctrl+C` in the terminal running the server.

## Troubleshooting

- If the browser cannot connect, check that the server is still running.
- If port `8083` is already in use, stop the other program or change `address` in `main.go`.
- If the page loads without styling, check that `index.html` links to `styles.css`.
- This example is for local development and learning, not production hosting.
