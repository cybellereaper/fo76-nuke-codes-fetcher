# Fallout 76 Nuke Codes Fetcher

This Go program fetches the latest Fallout 76 nuke codes from the [Fallout Builds website](https://www.falloutbuilds.com/fo76/nuke-codes/), parses them, and outputs the result as a formatted JSON object. It utilizes a retry mechanism in case of network failures or timeouts and routes traffic through the Tor network via a SOCKS5 proxy.

## Features

- **Fetches Fallout 76 Nuke Codes:** Retrieves the Alpha, Bravo, and Charlie silo codes.
- **Valid Time Information:** Extracts the validity periods for the nuke codes.
- **Retry Logic:** The program retries fetching the document up to 5 times in case of errors like timeouts or invalid status codes.
- **Tor Proxy Support:** Routes HTTP traffic through the Tor network using a SOCKS5 proxy.

## Prerequisites

- **Go**: This program requires Go version 1.23 or later.
- **Tor**: Tor must be running on your local machine and accessible via the SOCKS5 proxy (`127.0.0.1:9050`).
- **jq**: A command-line JSON processor is required for parsing the output (optional for manual inspection).

