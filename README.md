![Go](https://github.com/tryanzu/anzu/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tryanzu/core)](https://goreportcard.com/report/github.com/tryanzu/core)

# What is Anzu? 

Anzu is an open source software project designed to create communities. Simple, reactive, and performant forums software. 

This repository contains the core backend code and the frontend as a git submodule inside static/frontend. 

The system is a WIP and while usable things might change dramatically.

## 
![Anzu alpha post page screenshot](https://imgur.com/pXDutG0.png)

## Anzu's stack
- [Go](https://golang.org/) programming language
- [Redis](https://redis.io/) (required) for cache
- [MongoDB](https://www.mongodb.com/) (required)
- [React](https://reactjs.org/) for the webapp

## Installation

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
Start a local MongoDB 3.1 server with the help of docker and docker compose, ensure you have docker installed and then run:
`docker compose up`

### Last steps

Building the frontend before getting started is required, to do so, execute `npm install && npm run build` inside `static/frontend` submodule.
Once the frontend is built you can build the backend program with `go build -o anzu` and then execute `./anzu api` to run anzu's http web server.

If you are running anzu for the first time you should be able to log-in with the credentials:
```
username: admin@local.domain 
password: admin
```

## Commits

We follow the [Conventional Commits](https://www.conventionalcommits.org) specification, which help us with automatic semantic versioning and CHANGELOG generation.

## Contributing

We welcome contributions from the community! Whether it's reporting bugs, suggesting new features, or submitting code changes, your input is valuable. To get started:

1. Fork the repository and create a new branch for your contribution.
2. Make your changes and ensure they follow our coding style and guidelines.
3. Write clear commit messages following the [Conventional Commits](https://www.conventionalcommits.org) specification.
4. Test your changes thoroughly.
5. Submit a pull request with a detailed description of your changes.

We appreciate your help in making Anzu better! If you have any questions or need assistance, feel free to reach out to the maintainers or join our community channels.