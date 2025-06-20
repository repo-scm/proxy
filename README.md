# proxy

[![Build Status](https://github.com/repo-scm/proxy/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/repo-scm/proxy/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/repo-scm/proxy)](https://goreportcard.com/report/github.com/repo-scm/proxy)
[![License](https://img.shields.io/github/license/repo-scm/proxy.svg)](https://github.com/repo-scm/proxy/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/repo-scm/proxy.svg)](https://github.com/repo-scm/proxy/tags)



## Introduction

git sites proxy



## Usage

### 1. Run proxy server

```bash
# http://localhost:9090/ui
proxy serve [--address string]
```

### 2. Query available site

```bash
proxy query [--output string] [--verbose]
```

### 3. List available sites

```bash
proxy list
```



## Settings

[proxy](https://github.com/repo-scm/proxy) parameters can be set in the directory `$HOME/.repo-scm/proxy.yaml`.

An example of settings can be found in [proxy.yaml](https://github.com/repo-scm/proxy/blob/main/config/proxy.yaml).

```yaml
gerrits:
  - site_name: "your_site"
    api_base: "http://localhost:8080"
    user_name: "your_name"
    user_pass: "your_pass"
```



## License

Project License can be found [here](LICENSE).



## Reference
