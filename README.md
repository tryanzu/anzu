![Go](https://github.com/tryanzu/anzu/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tryanzu/core)](https://goreportcard.com/report/github.com/tryanzu/core)

<div align="center">
  <h1>🏛️ Anzu</h1>
  <p><strong>Modern, reactive community platform built for the next generation</strong></p>

![Anzu alpha post page screenshot](https://imgur.com/pXDutG0.png)

</div>

## ✨ Features

- **Real-time Communication**: WebSocket-powered live discussions and notifications
- **Gamification**: Built-in ranking system and user progression
- **Trust Network**: Advanced user trust calculation for content moderation
- **Modular Architecture**: Clean separation of concerns with dependency injection
- **Modern Frontend**: React-based SPA with responsive design
- **Rich Content**: Support for media uploads, markdown, and emoji reactions
- **OAuth Integration**: Google and Facebook authentication
- **Admin Tools**: Comprehensive administrative interface and CLI tools

## 🎯 What is Anzu?

Anzu is a modern, open-source community platform designed to foster engaging online communities. Built with Go and React, it combines the performance of modern backend architecture with a sleek, reactive frontend experience.

Whether you're building a gaming community, developer forum, or general discussion platform, Anzu provides the tools you need with built-in gamification, real-time features, and advanced moderation capabilities.

> ⚠️ **Early Development**: Anzu is in active development. While functional, the API and features may change significantly.


## 🚀 Deploy on Hostinger

[![Deploy on Hostinger](https://assets.hostinger.com/vps/deploy.svg)](https://www.hostinger.com/vps/docker-hosting?compose_url=https://github.com/tryanzu/anzu/)

## 🚀 Deploy to DigitalOcean

Deploy Anzu to the DigitalOcean App Platform with one click using the button below.

[![Deploy to DigitalOcean](https://www.deploytodo.com/do-btn-blue.svg)](https://cloud.digitalocean.com/apps/new?repo=https://github.com/tryanzu/anzu/tree/develop)

After deployment, configure the required environment variables in the DigitalOcean App Platform dashboard. Refer to the [Environment Variables](#-environment-variables) section for detailed configuration instructions.

**Required Services**: Your deployed app must be connected to both a Redis/Valkey database and a MongoDB instance to function properly.

## 🛠️ Tech Stack

- **Backend**: [Go 1.23+](https://golang.org/) with dependency injection
- **Frontend**: [React](https://reactjs.org/) SPA with Webpack
- **Database**: [MongoDB 8](https://www.mongodb.com/) for primary storage
- **Cache**: [Redis](https://redis.io/) for sessions and caching
- **Storage**: [MinIO](https://min.io/) S3-compatible object storage
- **Real-time**: WebSocket via custom Glue implementation

## 🚀 Quick Start

### Download dependencies

The first step is to download Go, official binary distributions are available at [https://golang.org/dl/](https://golang.org/dl/).

Now you need to download and configure **MongoDB** and **Redis**. Alternatively you can use remote servers.

### Download the repositories

Download the [core](http://github.com/tryanzu/anzu) in any path.

Initialize the repo submodule, so the [frontend](http://github.com/tryanzu/frontend) is in `static/frontend`.

```
git submodule update --init --recursive
```

### Configure

Copy the `.env.example` file into `.env` and edit it to meet your local environment configuration.

### Run MongoDB with docker compose

Start a local MongoDB 8 server with the help of docker and docker compose, ensure you have docker installed and then run:
`docker compose up`

### Last steps

Building the frontend before getting started is required, to do so, execute `npm install && npm run build` inside `static/frontend` submodule.
Once the frontend is built you can build the backend program with `go build -o anzu ./cmd/anzu` and then execute `./anzu api` to run anzu's http web server.

If you are running anzu for the first time you should be able to log-in with the credentials:

```
username: admin@local.domain
password: admin
```

## 🔧 Development

### Available Commands

```bash
# Backend Development
go build -o anzu ./cmd/anzu   # Build the backend
./anzu api                    # Start API server (port 3200)
./anzu shell                  # Interactive admin shell
./anzu sync-ranking           # Sync gaming rankings

# Frontend Development
cd static/frontend
npm install                   # Install dependencies
npm start                     # Development build with watch
npm run build                 # Production build
npm run eslint                # Lint JavaScript

# Code Quality
golangci-lint run             # Lint Go code
golangci-lint run --fast      # Quick lint checks

# Services
docker compose up             # Start MongoDB, MinIO, mongo-express
```

### Project Structure

```
anzu/
├── board/                    # Board domain logic
│   ├── events/              # Event handlers
│   ├── posts/               # Post management
│   ├── comments/            # Comment system
│   └── votes/               # Voting system
├── modules/                 # Core modules
│   ├── api/                 # HTTP API endpoints
│   ├── user/                # User management
│   ├── gaming/              # Gamification
│   └── acl/                 # Access control
├── core/                    # Core services
│   ├── config/              # Configuration
│   ├── events/              # Event system
│   └── content/             # Content processing
└── static/frontend/         # React frontend
```

## 📋 Commits

We follow the [Conventional Commits](https://www.conventionalcommits.org) specification, which help us with automatic semantic versioning and CHANGELOG generation.

## 🏗️ Architecture

Anzu follows a modular, event-driven architecture with clear separation of concerns:

### Core Principles

- **Dependency Injection**: Uses Facebook's inject library for clean DI
- **Event-Driven**: Centralized event handling for cross-module communication
- **Modular Design**: Self-contained modules with clear interfaces
- **Trust Network**: User trust calculation system for content moderation

### Key Modules

- **Board Domain** (`board/`): Posts, comments, votes, and content management
- **User Module** (`modules/user/`): Authentication, profiles, OAuth integration
- **Gaming Module** (`modules/gaming/`): Ranking system and gamification
- **ACL Module** (`modules/acl/`): Role-based access control and permissions
- **API Module** (`modules/api/`): HTTP endpoints and REST API using Gin

### Real-time Features

- WebSocket communication via custom Glue implementation
- Live notifications and real-time discussions
- Event-driven updates across the platform

## 🤝 Contributing

We welcome contributions from the community! Whether it's reporting bugs, suggesting new features, or submitting code changes, your input is valuable.

### Getting Started

1. Fork the repository and create a new branch for your contribution
2. Make your changes following our coding style and guidelines
3. Write clear commit messages using [Conventional Commits](https://www.conventionalcommits.org)
4. Test your changes thoroughly using the development commands above
5. Submit a pull request with a detailed description

### Development Workflow

- Run `golangci-lint run` before submitting Go code
- Run `npm run eslint` for frontend changes
- Use `go build -o anzu ./cmd/anzu && ./anzu api` for backend development
- Test with the Docker Compose environment

We appreciate your help in making Anzu better! If you have questions, feel free to open an issue or reach out to the maintainers.

## 📚 Additional Resources

### API & Documentation

- **API Server**: Runs on `http://localhost:3200` by default
- **Admin Panel**: Access via web interface with admin credentials
- **MongoDB Admin**: Mongo Express available at `http://localhost:8081`
- **MinIO Console**: S3 storage admin at `http://localhost:9000`

### Configuration

- **Environment**: Copy `.env.example` to `.env` and customize
- **Database**: MongoDB connection configured via `MONGO_URL`
- **Storage**: S3-compatible storage via MinIO or AWS S3
- **Authentication**: JWT tokens with OAuth support (Google, Facebook)

### Community

- **Issues**: Report bugs and request features on GitHub
- **Discussions**: Join community discussions and get help
- **Wiki**: Additional documentation and guides (coming soon)

### Environment Variables

When deploying to DigitalOcean, you will need to configure the following environment variables in the App Platform dashboard.

| Variable                       | Type     | Required | Description                                                                        |
| ------------------------------ | -------- | -------- | ---------------------------------------------------------------------------------- |
| `JWT_SECRET`                   | `Secret` | Yes      | A long, random string used to sign JWT tokens.                                     |
| `S3_ENDPOINT`                  | `String` | Yes      | The endpoint for your S3-compatible storage (e.g., `nyc3.digitaloceanspaces.com`). |
| `S3_ACCESS_KEY_ID`             | `Secret` | Yes      | The access key for your S3-compatible storage.                                     |
| `S3_SECRET_ACCESS_KEY`         | `Secret` | Yes      | The secret key for your S3-compatible storage.                                     |
| `S3_BUCKET_NAME`               | `String` | Yes      | The name of the S3 bucket to use for asset storage.                                |
| `S3_REGION`                    | `String` | Yes      | The region where your S3 bucket is located.                                        |
| `OAUTH_GOOGLE_CLIENT_ID`       | `Secret` | No       | The client ID for Google OAuth.                                                    |
| `OAUTH_GOOGLE_CLIENT_SECRET`   | `Secret` | No       | The client secret for Google OAuth.                                                |
| `OAUTH_FACEBOOK_CLIENT_ID`     | `Secret` | No       | The client ID for Facebook OAuth.                                                  |
| `OAUTH_FACEBOOK_CLIENT_SECRET` | `Secret` | No       | The client secret for Facebook OAuth.                                              |
| `SMTP_HOST`                    | `String` | No       | The hostname of your SMTP server.                                                  |
| `SMTP_PORT`                    | `String` | No       | The port of your SMTP server.                                                      |
| `SMTP_USERNAME`                | `Secret` | No       | The username for your SMTP server.                                                 |
| `SMTP_PASSWORD`                | `Secret` | No       | The password for your SMTP server.                                                 |
| `SMTP_FROM_EMAIL`              | `String` | No       | The email address to send emails from.                                             |
| `SENTRY_URL`                   | `Secret` | No       | The DSN for Sentry error tracking.                                                 |

---

<div align="center">
  <p>Built with ❤️ by the Anzu community</p>
  <p>
    <a href="https://github.com/tryanzu/anzu/issues">Report Bug</a> ·
    <a href="https://github.com/tryanzu/anzu/issues">Request Feature</a> ·
    <a href="https://github.com/tryanzu/anzu/discussions">Discussions</a>
  </p>
</div>
