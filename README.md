![Go](https://github.com/tryanzu/anzu/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tryanzu/core)](https://goreportcard.com/report/github.com/tryanzu/core)

# Meet Anzu

Anzu is our greatest endeavor to build the most rad, simple & reactive forum software out there since the Javascript revolution. 

Forum platforms to host communities are vast. Many would say it's a lifeless space with almost zero innovation, and attempting to create something new is pointless. We dissent, and if you found this repository you might also share with us the idea that there has to be an alternative to the old forum. Well, we think Anzu is that young and sexy software that could bring back to life the community-building movement.

This repository contains the core backend and the frontend submodule link. 
We're still working in the first alpha, so previous knowledge about the stack is required to set things up.

## Alpha screenshots
![Anzu alpha post](https://imgur.com/pXDutG0.png)
![Anzu alpha publisher](https://imgur.com/tF1ApnP.png)
![Anzu alpha post](https://imgur.com/IAv9V8C.png)
![Anzu alpha chat](https://imgur.com/vlari7x.png)
![Anzu alpha profile](https://imgur.com/uG4C9LE.png)

## Anzu's stack
- [Go](https://golang.org/) programming language
- [redis](https://redis.io/) (required) for cache
- [mongoDB](https://www.mongodb.com/) (required)
- [react](https://reactjs.org/) for the webapp

# Contribute

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

Install the core dependencies with `go build`.

Install the frontend dependencies with `yarn install`.


### Configure

Copy the `.env.example` file into `.env` and edit it to meet your local environment configuration.

### Last steps

Building the frontend before getting started is required, to do so, execute `npm install && npm run build` inside `static/frontend` submodule.
Once the frontend is built you can build the backend program with `go build -o anzu` and then execute `./anzu api` to run anzu's http web server.

## Commits

We follow the [Conventional Commits](https://www.conventionalcommits.org) specification, which help us with automatic semantic versioning and CHANGELOG generation.