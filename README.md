# Snarky
### Zero-Knowledge, Asynchronous Dead Drop

![Snarky Logo](https://sapadian.com/assets/snarky_logo.png)

**Snarky** is a self-hosted, RAM-only file transfer tool designed for DevOps teams and Sysadmins. It solves the problem: *"I need to send you a sensitive secret (password, key, config), but you aren't online, and I don't want it stored in our chat logs."*

---

## 🚀 Features

* **Zero-Knowledge:** Encryption happens on the client *before* upload. The server never sees the encryption key.
* **Burn-After-Reading:** The moment a file is downloaded, it is deleted from the server instantly.
* **RAM-Only Storage:** Data is stored in the server's volatile memory. If the power is cut, the data vanishes.
* **Ephemeral:** Files not picked up within 24 hours are automatically incinerated.

---

## 📦 Client Installation

You can download pre-compiled binaries from the [Releases Page](https://sapadian.com/snarky/clients) or build it yourself.

### Build form Source
```bash
git clone [https://github.com/sapadian/snarky.git](https://github.com/sapadian/snarky.git)
cd snarky
go build -o snarky main.go
