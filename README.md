# tsuru-api-docs

a tool to extract docs from handler comments

## Installation

`tsuru-api-docs` requires Go 1.6 or later.

```
go get -u github.com/tsuru/tsuru-api-docs
```

## Usage

### Print a yaml with the api docs

```
tsuru-api-docs
```

### Check handlers without docs

```
tsuru-api-docs | grep missing
```
