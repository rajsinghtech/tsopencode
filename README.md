# tsopencode

Run [opencode](https://opencode.ai) on your Tailscale tailnet. No system `tailscaled` required — embeds Tailscale via [tsnet](https://pkg.go.dev/tailscale.com/tsnet).

## Install

```bash
brew install rajsinghtech/tap/tsopencode
```

## Usage

```bash
# First run — authenticate once (browser or authkey)
export TS_AUTHKEY=tskey-auth-...   # optional, falls back to browser login
tsopencode

# Subsequent runs — fully headless (state persisted)
tsopencode

# Custom node name
tsopencode --hostname mydevbox
# → https://mydevbox.<tailnet>.ts.net
```

## Flags

| Flag | Env | Default | Description |
|---|---|---|---|
| `--authkey` | `TS_AUTHKEY` | — | Tailscale auth key |
| `--hostname` | `TSOPENCODE_HOSTNAME` | `opencode` | Tailscale node name |
| `--state-dir` | `TSOPENCODE_STATE_DIR` | `~/.config/tsopencode/` | State base dir |
| `--opencode-bin` | — | `opencode` | Path to opencode binary |

State is stored in `<state-dir>/tsnet-state/` and persists across runs.
