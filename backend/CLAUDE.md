# Claude Code Instructions for Dev-Share Backend

## Project Overview

This is a Go backend service for managing developer environments. Clients can provision, configure, and control their development infrastructure through this API.

## Architecture Principles

- **Clean Architecture**: Follow separation of concerns with clear boundaries between layers
- **Domain-Driven Design**: Domain logic lives in `internal/domain/`
- **Dependency Rule**: Dependencies point inward toward the domain layer

## Technology Stack

- **Web Framework**: Fiber v2 - Use for all HTTP routing and middleware
- **Entry Point**: `cmd/server/main.go` - Main application entry point
- **Database Migrations**: golang-migrate/migrate - All migrations in `internal/infra/migrations/`
  - See `.claude/rules/database-migrations.md` for detailed migration guidelines

## Code Style Guidelines

### Go Conventions
- Follow standard Go idioms and conventions
- Use `gofmt` and `go vet` for formatting and linting
- Prefer small, focused functions with clear names
- Error handling: always check and handle errors explicitly, don't ignore them
- Use meaningful variable names; avoid single-letter names except for short-scope iterators
