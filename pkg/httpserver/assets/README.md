# Server Help

This is a small local HTTP server. It serves the example page, a health check, and embedded fallback pages when project files are missing.

## Start The Server

Windows:

```powershell
server.exe
```

Linux or macOS:

```bash
./server
```

## Check If It Is Running

```bash
curl http://localhost:8081/health
```

Expected response:

```text
OK
```

You can also open:

```text
http://localhost:8081/
```

## Edit The Site

Edit the HTML and CSS files in the project:

- `web/templates/index.html`
- `web/static/styles.css`

## Check Config

Look at `server.json` near the executable or in the working directory.

Common fields:

- `server_address`
- `template_dir`
- `static_dir`
- `allowed_hosts`
- `log_file_name`
- `log_level`

## Check Logs

Look at the configured log file, usually `app.log`.

Logs can show startup problems, missing files, rejected hosts, and panic recovery messages.

## Common Problems

### Port Already In Use

Another program may already be using the port.

Windows:

```powershell
netstat -ano | findstr :8081
```

Linux or macOS:

```bash
lsof -i :8081
```

Fix: stop the other program or change `server_address` in `server.json`.

### Health Check Fails

If this does not return `OK`:

```bash
curl http://localhost:8081/health
```

Check that:

- the server is still running,
- the port matches `server_address`,
- the log file does not show a startup error.

### Built-In Page Appears

The embedded page is a fallback. It usually means the template file is missing or the template folder is misconfigured.

Check:

- `template_dir` in `server.json`
- `web/templates/index.html`
- the log file for missing template messages

### Styles Do Not Load

Check:

- `static_dir` in `server.json`
- `web/static/styles.css`
- the browser URL for the stylesheet

