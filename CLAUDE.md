# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

**go-docgen** — Go library for generating documents from templates (PDF from HTML, DOCX, and related formats). Not an HTTP microservice; consumers import this module.

## Commands

```bash
go mod tidy
go build ./...
go test ./...
```

## Conventions

- Library code lives at repo root and under `cmd/pdfcompare/` (separate nested module).
- Keep template assets and rendering logic separated from consumer apps.
- Add tests when changing template parsing or output formats.
