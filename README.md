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

### List handlers with a specific method

```
tsuru-api-docs -method GET
```

### List handlers without a specific method

```
tsuru-api-docs -no-method GET
```

### List handlers matching search regexp

```
tsuru-api-docs -search "event\.New"
```

### List handlers NOT matching search regexp

```
tsuru-api-docs -no-search "event\.New"
```
