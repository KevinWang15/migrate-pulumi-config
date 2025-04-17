# migrate-pulumi-config

Copy **all** Pulumi stack config values—secrets included—from one stack to another in a single command.

> Re‑encrypts secrets on the fly using the destination stack’s passphrase.

---

## ✨ Features

| Feature | Details |
|---------|---------|
| ⚡ **Fast**  | Shells out to the Pulumi CLI only for the exact commands needed. |
| 🔐 **Secure** | Secrets are never written to disk. Everything happens in‑memory. |
| 🚦 **Idempotent** | Skip‑creates the destination stack if it doesn’t exist. |
| 🤝 **Plain Go** | No CGO, external deps, or Pulumi automation SDK required. |

---

## 📦 Installation

```bash
# Clone and install the tool (adds it to your $GOBIN)
git clone https://github.com/KevinWang15/migrate-pulumi-config.git
cd migrate-pulumi-config
# Build & install
GO111MODULE=on go install .
```

> **Tip:** make sure `$GOBIN` (default `~/go/bin`) is on your `PATH`.

---

## 🚀 Quick Start

Run the utility *inside* the folder that contains your Pulumi project files (`Pulumi.yaml`, `Pulumi.<stack>.yaml`).

```bash
# Inside your Pulumi project directory
migrate-pulumi-config \
  --src dev \            # source stack name (NOT a file name)
  --dst dev-new \        # destination stack (created if missing)
  --src-pass "oldPass" \
  --dst-pass "newPass"
```

### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--src` | ✅ | Source stack name (e.g., `dev`) |
| `--dst` | ✅ | Destination stack name |
| `--src-pass` | ✅ | Passphrase for the **source** stack |
| `--dst-pass` | ✅ | Passphrase for the **destination** stack |
| `--dir` | 🚫 | Pulumi project directory (default `.`) |

---

## 🔍 What it does

1. **Decrypts** all config values from the source stack using `pulumi config --show-secrets`.
2. Ensures the destination stack exists (`pulumi stack init` if needed).
3. **Re‑sets** each key on the destination stack—adding `--secret` when the original value was a secret—so Pulumi encrypts it using the *new* passphrase.

---

## 🛠  Building from source

```bash
go build -o migrate-pulumi-config ./
```

---

## 📄 License

MIT