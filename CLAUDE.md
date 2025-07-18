# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Running
- **Build backend**: `go build -o anzu` (creates executable)
- **Run API server**: `./anzu api` (starts HTTP server on port 3200)
- **Development mode**: `go build -o anzu && ./anzu api` (manual restart required)
- **Interactive shell**: `./anzu shell` (maintenance and admin tools)

### Frontend Development
```bash
cd static/frontend
npm install          # Install dependencies
npm run build        # Production build
npm start            # Development build with watch
npm run eslint       # Lint JavaScript code
```

### Code Quality
- **Lint Go code**: `golangci-lint run` (install with `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- **Quick lint**: `golangci-lint run --fast` (faster checks only)

### Database and Services
- **Start MongoDB**: `docker compose up` (includes MongoDB 8, mongo-express, and MinIO)
- **Sync ranking**: `./anzu sync-ranking` (recalculates gaming rankings)

## Architecture Overview

### Backend Structure
- **Go 1.23+ application** using dependency injection (facebookgo/inject)
- **Modular architecture** with clear separation of concerns
- **Event-driven system** with centralized event handlers in `board/events/`

### Key Components

#### Core Modules (`modules/`)
- **API**: HTTP router and REST endpoints (Gin framework)
- **User**: Authentication, profiles, and user management
- **Gaming**: Ranking system and gamification features
- **ACL**: Permission and role-based access control
- **Security**: Trust network and user trust calculations
- **Notifications**: Email and in-app notification system
- **Feed**: Post aggregation and activity feeds

#### Board Domain (`board/`)
- **Posts**: Content creation and management
- **Comments**: Threaded discussions
- **Categories**: Content organization
- **Votes**: Reaction system (upvotes/downvotes)
- **Flags**: Content moderation
- **Realtime**: WebSocket communication using Glue

#### Core Services (`core/`)
- **Config**: Application configuration management
- **Events**: Centralized event handling system
- **HTTP**: Middleware and HTTP utilities
- **Content**: Text processing and mention parsing

### Frontend
- **React-based SPA** in `static/frontend/`
- **Webpack build system** with development and production configs
- **SCSS theming** with multiple theme support
- **Real-time features** via WebSocket integration

### Database
- **MongoDB** as primary database
- **Redis** for caching and sessions
- **MinIO** for S3-compatible object storage

### Key Patterns
- **Dependency Injection**: Uses Facebook's inject library for DI
- **Event System**: Centralized event handling for cross-module communication
- **Module Pattern**: Each feature is a self-contained module with deps, finders, model, and mutators
- **Trust Network**: User trust calculation system for content moderation

### Authentication
- **JWT tokens** for API authentication
- **OAuth integration** (Google, Facebook) via Goth library
- **Session-based** web authentication

### Development Notes
- Default admin credentials: `admin@local.domain` / `admin`
- Uses Cobra for CLI commands
- Follows Conventional Commits specification
- Real-time features via `/glue/` WebSocket endpoint