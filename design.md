# Go Backend Assignment – Wallet Service

## Overview

Build a simple wallet service in Go with a REST API.

This assignment is expected to take 2–4 hours. You may use any libraries or references. The assignment should be written in Go.

---

## Functional Requirements

### 1. Create Wallet

`POST /wallets`

- Returns a unique wallet ID
- Initial balance is zero

### 2. Get Wallet

`GET /wallets/{wallet_id}`

- Returns wallet ID and current balance

### 3. Transfer Funds

`POST /wallets/transfer`

- Transfers an amount from one wallet to another
- Request body must include source wallet, destination wallet, and amount

---

## Technical Requirements

- Language: Go
- Storage: in-memory only, no external databases
- Service must start with: `go run ./cmd/server`

---

## Deliverables

Submit a **public GitHub repository** containing:

- Complete source code with commit history
- A `README.md` with instructions to run and test the service

> Repositories submitted as a zip file or without meaningful commit history will not be reviewed.

---

## Bonus

- gRPC API (sharing the same business logic as REST)
- Docker support
- Documentation
- Load/concurrency testing script or tool
- External Database
- Horizontal scaling

---

## Notes

- The service does not need to be production-ready
- Prefer simplicity and correctness over complexity
- Incomplete or non-runnable submissions will not be reviewed