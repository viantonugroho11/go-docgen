# Contributing to go-docgen

Thank you for your interest in contributing to **go-docgen** 🎉
We welcome contributions of all kinds: bug reports, feature requests, documentation improvements, and code.

---

## 📌 How to Contribute

### 1. Fork the Repository

Create your own fork of the repository and clone it locally:

```bash
git clone https://github.com/<your-username>/go-docgen.git
cd go-docgen
```

---

### 2. Create a Branch

Use a clear and descriptive branch name:

```bash
git checkout -b feature/add-csv-streaming
```

Examples:

* `feature/...`
* `fix/...`
* `docs/...`
* `refactor/...`

---

### 3. Make Your Changes

* Follow existing code structure and style
* Keep functions small and focused
* Add comments where necessary
* Avoid breaking public APIs unless discussed first

---

### 4. Add Tests

If applicable:

* Add unit tests for new features
* Ensure existing tests still pass

Run tests:

```bash
go test ./...
```

---

### 5. Run Benchmarks (Optional but Recommended)

```bash
go test -run='^$' -bench . -benchmem ./...
```

This helps ensure performance is not degraded.

---

### 6. Commit Your Changes

Use clear commit messages:

```bash
git commit -m "feat: add CSV streaming support"
```

Recommended format:

* `feat:` new feature
* `fix:` bug fix
* `docs:` documentation
* `refactor:` internal changes
* `test:` test updates

---

### 7. Push & Create Pull Request

```bash
git push origin your-branch-name
```

Then open a Pull Request (PR) to the `main` branch.

---

## ✅ Pull Request Guidelines

* Keep PRs small and focused
* Provide a clear description:

  * What changed?
  * Why it’s needed?
  * Any breaking changes?
* Link related issues (if any)

---

## 🧪 Code Style Guidelines

* Use idiomatic Go
* Follow `gofmt` and `go vet`
* Prefer composition over inheritance
* Keep packages loosely coupled

---

## 📦 Project Structure Overview

```text
export.go          → public API (facade)
engine/            → output engines (PDF, CSV, Excel)
template/          → template rendering
internal/loader/   → file loading & caching
```

---

## 🚫 What to Avoid

* Large, unrelated changes in a single PR
* Breaking public API without discussion
* Adding heavy dependencies without strong reason

---

## 💡 Suggestions & Ideas

If you have ideas (e.g. new export format, performance improvements, template system enhancements):

* Open an issue first
* Discuss design before implementation

---

## 🙌 Code of Conduct

Be respectful and constructive.
We aim to maintain a welcoming and collaborative environment.

---

## 🚀 Maintainers

Maintained by the go-docgen contributors.

---

Thanks again for contributing! 💙
