# MCP Toolbox

## Setting up
Check release version from https://github.com/googleapis/genai-toolbox/releases

```
go install github.com/googleapis/genai-toolbox@v0.14.0
```

Run

```
genai-toolbox --ui --tools-file tools.yaml --log-level DEBUG
```

## Configuration

```
sources:
  hotel-source:
    kind: http
    baseUrl: https://gist.githubusercontent.com/nanikjava/761db1bbe779d1f675f870013da5896a/raw/93fa3810c5740db6d4b8ec9660be45eff311c4d2/hotels.txt
    timeout: 10s # default to 30s

tools:
    search-hotels-by-name:
        kind: http
        source: hotel-source
        method: GET
        path: /
        description: Tool to update information to the example API

toolsets:
    hotel:
        - search-hotels-by-name
```

# Run example

```
GEMINI_API_KEY=<api_key> go run main.go
```