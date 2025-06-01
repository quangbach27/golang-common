# golang-common

A Go module providing common utilities and middleware for Go web applications.

## Features

- HTTP routing with [chi](https://github.com/go-chi/chi)
- CORS middleware
- Structured logging with [logrus](https://github.com/sirupsen/logrus)
- Prefixed log formatting

## Requirements

- Go 1.24.3 or later

## Dependencies

- github.com/go-chi/chi/v5
- github.com/go-chi/cors
- github.com/sirupsen/logrus
- github.com/x-cray/logrus-prefixed-formatter

## Getting Started

Clone the repository:

```sh
git clone https://github.com/quangbach27/golang-common.git
cd golang-common
```

Add the repository:

```sh
go get github.com/quangbach27/golang/common
```

Import and use it in your code:

```sh
import "github.com/quangbach27/golang/common"
```
