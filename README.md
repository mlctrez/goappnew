# goappnew

## Purpose

A [go-app](https://go-app.dev/) project structure to bootstrap new projects or experiments.

This is intended to be used with [gonew](https://go.dev/blog/gonew) and
replaces [goappcreate](https://github.com/mlctrez/goappcreate) which did basically the same copy/substitute operations.

The following features are included:

* A Makefile supporting all the build steps required to run a [go-app](https://go-app.dev/) project.
* An application version backed by the current tag and commit hash from git.
* A proper split between the wasm and server dependencies to reduce the final wasm size.
* Web resources, including the `app.wasm` are embedded in the server binary during the build.
* The server binary uses [servicego](github.com/mlctrez/servicego) to allow installation as a service on
  supported platforms.
* The [gin web framework](https://github.com/gin-gonic/gin) is included and brotli compression middleware
  is pre-configured.
* It supports the wasm file size header to correctly display progress on the loading screen when using compression.
* The `app.Handler` configuration is loaded from a json file in the web directory.
* When running in dev mode, [app updates](https://go-app.dev/lifecycle#listen-for-app-updates) are automatic.
* Uses just a single folder `goapp` so it can be copied into an existing codebase.
* Less than 300 lines of code - easy to understand and modify.

## Usage

* gonew github.com/mlctrez/goappnew@v0.2.0 github.com/mlctrez/goappnewtest 
* cd goappnewtest
* Run `make` with no arguments to start the server in dev mode.

created by [tigwen](https://github.com/mlctrez/tigwen)
