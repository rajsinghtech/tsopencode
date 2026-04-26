# tsopencode

Run [opencode](https://opencode.ai) on your Tailscale tailnet. No system `tailscaled` required — embeds Tailscale via [tsnet](https://pkg.go.dev/tailscale.com/tsnet).

Once running, opencode is accessible at `https://<hostname>.<tailnet>.ts.net` and `http://<hostname>.<tailnet>.ts.net` from any device on your tailnet.

## Install

```bash
brew install rajsinghtech/tap/tsopencode
```

## Quick start

```bash
# Authenticate once (browser login if no authkey)
export TS_AUTHKEY=tskey-auth-...
tsopencode
# → https://opencode.your-tailnet.ts.net

# Check status at any time
tsopencode status
```

## Running as a background service

```bash
brew services start tsopencode   # start at login
brew services stop tsopencode    # stop
brew services restart tsopencode # restart after upgrade
```

Logs go to `/opt/homebrew/var/log/tsopencode.log`.

## Commands

| Command | Description |
|---|---|
| `tsopencode` | Start the server (foreground) |
| `tsopencode status` | Show running state, URL, logs, and state dir |

## Flags

| Flag | Env | Default | Description |
|---|---|---|---|
| `--authkey` | `TS_AUTHKEY` | — | Tailscale auth key for headless registration |
| `--hostname` | `TSOPENCODE_HOSTNAME` | `opencode` | Tailscale node name |
| `--state-dir` | `TSOPENCODE_STATE_DIR` | `~/.config/tsopencode/` | Base dir for tsnet state |
| `--opencode-bin` | — | `opencode` | Path to opencode binary |

Tailscale node state persists to `<state-dir>/tsnet-state/` — subsequent runs are fully headless once authenticated.

## How it works

tsopencode spawns `opencode serve` on a random local port, brings up a tsnet node (its own Tailscale identity, no system daemon required), and reverse-proxies all HTTP and WebSocket traffic from the tailnet to the local opencode process.

```
your devices  ──►  tsnet :443 / :80  ──►  reverse proxy  ──►  opencode serve
  (tailnet)                                                      (127.0.0.1:N)
```

opencode uses your global config (`~/.config/opencode/`) for API keys and provider settings. Projects are resolved per-workspace at connection time.
