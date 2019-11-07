#  Meet Anzu

  

Anzu is our greatest endeavor to build the most rad, simple & reactive forum software out there since the Javascript revolution.

Forum platforms to host communities are vast. Many would say it's a lifeless space with almost zero innovation, and attempting to create something new is pointless. We dissent, and if you found this repository you might also share with us the idea that there has to be an alternative to the old forum. Well, we think Anzu is that young and sexy software that could bring back to life the community-building movement.

This repository contains the front-end repository.

We're still working in the first alpha, so previous knowledge about the stack is required to set things up.

##  Alpha screenshot

![Anzu alpha post](https://imgur.com/pXDutG0.png)
![Anzu alpha publisher](https://imgur.com/tF1ApnP.png)
![Anzu alpha post](https://imgur.com/IAv9V8C.png)
![Anzu alpha chat](https://imgur.com/vlari7x.png)
![Anzu alpha profile](https://imgur.com/uG4C9LE.png)

  

##  Anzu's stack

-  Golang.

-  Redis (to be replaced)

-  BuntDB (embedded cache)

-  MongoDB (DB)

-  Cycle.JS (A functional and reactive JavaScript framework)

  

#  Contribute

  

##  Installation

  

###  Download dependencies

The first step is to download Go, official binary distributions are available at [https://golang.org/dl/](https://golang.org/dl/).

If you are upgrading from an older version of Go you must first [remove the existing version](https://golang.org/doc/install?download=go1.11.4.darwin-amd64.pkg#uninstall).

  

[Download the package file](https://golang.org/dl/), open it, and follow the prompts to install the Go tools. The package installs the Go distribution to /usr/local/go.

The package should put the /usr/local/go/bin directory in your PATH environment variable. You may need to restart any open Terminal sessions for the change to take effect.

  

Make sure you have defined your GOPATH:

  

```zsh

export GOPATH=$HOME/go

export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

```

Download and configure **MongoDB** and **Redis**. You'll need to create a root user in MongoDB
In database>uri set your MongoDB user and password

Alternatively you can use remote servers.

  

Install `dep` for go dependencies: [https://github.com/golang/dep](https://github.com/golang/dep). In MacOS it can be installed with `brew`.

  

Execute the following command: `go get https://github.com/cespare/reflex`

  

Reflex probably only works on Linux and Mac OS.

  

###  Download the repositories

  

Download the [core](http://github.com/tryanzu/core) in any path.

Initialize the repo submodule, so the [frontend](http://github.com/tryanzu/frontend) is in `static/frontend`.

```

git submodule update --init --recursive

```
Install the core dependencies with `go build -o anzu`.

Install the frontend dependencies with `yarn install`.

###  Configure

 
Copy the `env.json.example` file into `env.json` and edit it to meet your local environment configuration.
In database>uri you'll need to set your MongoDB user and password
  
  Copy the `config.toml.example` file into `config.toml`

###  Last steps

Start mongo service `service mongod start`
  
Execute `./anzu`
[Execute](https://www.cyberciti.biz/faq/run-execute-sh-shell-script/) the script `develop.sh` to have hot reload (compile and run) while editing the core code.

Execute `npm run start` in `static/frontend` to have hot reload while editing the frontend code.

  

##  Commits

  

We follow the [Conventional Commits](https://www.conventionalcommits.org) specification, which help us with automatic semantic versioning and CHANGELOG generation.