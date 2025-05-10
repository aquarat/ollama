# Distributed Inferencing

Ollama supports distributed inferencing using RPC servers, allowing you to run models across multiple machines.

## Overview

Distributed inferencing allows you to:

- Distribute model layers across multiple machines
- Utilize different GPU types on different machines
- Scale up inferencing capabilities beyond a single machine's resources

## Running RPC Servers

Ollama now includes an integrated RPC server, eliminating the need to build and run a separate application. You can start an RPC server on any machine with:

```sh
ollama rpc-server
```

By default, the RPC server binds to `127.0.0.1:50052`. You can customize the host and port:

```sh
ollama rpc-server -H 0.0.0.0 -p 50053
```

For more details on the integrated RPC server, see [RPC Server Documentation](rpc_server.md).

> [!IMPORTANT]
> Never expose the RPC server to an open network or in a sensitive environment! This feature is experimental and not secure.

## Using RPC Servers with Ollama

### Using Environment Variable

The RPC servers can be specified using the `OLLAMA_RPC_SERVERS` environment variable.
This environment variable contains a comma-separated list of the RPC servers, e.g. `OLLAMA_RPC_SERVERS="127.0.0.1:50052,192.168.0.69:50053"`.
`ollama serve` will automatically offload the model to the RPC servers when this variable is set.

```sh
OLLAMA_RPC_SERVERS="127.0.0.1:50052,192.168.0.69:50053" ollama serve
```

### Change RPC Servers Using Request Options

The RPC servers can be changed using the `rpc_servers` option when generating a response.

```sh
curl http://localhost:11434/api/generate --json '{
  "model": "llama3.1",
  "prompt": "hello",
  "stream": false,
  "options": {
    "rpc_servers": "127.0.0.1:50053"
  }
}'
```

Ollama will use the RPC server `127.0.0.1:50053` instead of the servers set by the `OLLAMA_RPC_SERVERS` environment variable.

## Troubleshooting

If Ollama is having issues connecting to your RPC servers, make sure:

1. The RPC server is running and accessible from the Ollama server
2. There are no firewalls blocking the connection
3. The RPC server version matches the Ollama version
