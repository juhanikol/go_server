# Server Help

This executable starts a small HTTP server. It can serve project routes, a health check, and built-in fallback pages when project templates are missing.

## Start The Server

Windows:

```powershell
.\server.exe
```

macOS or Linux:

```bash
./server
```

If the server starts correctly, it should print startup log lines and listen on its configured address.

## Check If It Is Running

Default example address:

```bash
curl http://localhost:8081/health
```

Expected response:

```text
OK
```

You can also open this in a browser:

```text
http://localhost:8081/
```

## Check The Config File

Look for `server.json` near the executable or in the configured app folder.

Common values to check:

- `server_address`
- `log_file_name`
- `log_level`
- `template_dir`
- `static_dir`

For `log_level`, use a readable string such as:

```json
"log_level": "INFO"
```

## Check The Log File

Look for the configured log file, commonly:

```text
app.log
```

The log file can show startup errors, missing templates, route errors, and panic recovery messages.

## Common Problems

### Port Already In Use

If startup says the address is already in use, another program is using the port.

Windows PowerShell:

```powershell
netstat -ano | findstr :8081
```

macOS or Linux:

```bash
lsof -i :8081
```

Fix: stop the other program or change `server_address` in `server.json`.

### Health Check Fails

If this fails:

```bash
curl http://localhost:8081/health
```

Check that:

- the server executable is still running,
- the port matches `server_address`,
- a firewall is not blocking local connections,
- the log file does not show a startup error.

### Homepage Shows A Built-In Page

This usually means the project homepage template was not found or no custom homepage was configured.

Check:

- `template_dir` in `server.json`,
- whether the template folder exists,
- whether the log file mentions a missing template directory.

### Static Files Do Not Load

Check:

- `static_dir` in `server.json`,
- whether the static folder exists,
- whether the requested file path is correct.
