## Blockchain monitor

### Setup

Build using docker with:

    docker build . -t monitor

Then run with:

    docker run -p 8989:8989 monitor --node wss://<NODE ADDRESS>

It needs the websocket connection string to ensure it can subscribe to the event
stream from the node.

By default, it is running an HTTP server at port 8989 which exposed some basic
metrics and latest head block.

Access the `/head` endpoint for the last block observed (may take a while):

    % curl -D - http://localhost:8989/head
    HTTP/1.1 200 OK
    Content-Type: application/json
    Date: Mon, 26 Feb 2024 23:02:45 GMT
    Transfer-Encoding: chunked
    
    {
    "header": {
        "parentHash": "0x48c08d310951ab73722bad2678eb525f4ddfb583186950cf47f7d013210d9525",
        "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
        "miner": "0x4838b106fce9647bdf1e7877bf73ce8b0bad5f97",
        "stateRoot": "0xd6379d3277b6a79e712bf8b3946542f99a75e31b5f1d10aa97e265e6287677bc",
        "transactionsRoot": "0x392bbb7fd0697f46c2585ac1a3dfbc6bc6db78682a8a8ca96ff0fc3d8f1dde27",
        "receiptsRoot": "0x11818fbd40cedecee111370097f08d5af3c32c0048277e92fb8f5001cd51b51c",
        "logsBloom": "0x686b3b09c906e1e6505d93aaa706f29010e74ae0cc4eb4f5b6d9478c9c1985225a3325aa4ec053a0e31395ca2dd40b192ee3e3788b956d61c965b2a455ac2ca2485690d948235ebd8e8ac02d9802a1f9d06480c3945d5c7610444c50d66630b50c660d34ae55c802704ce5e002621c9d6a3bc83615ceffd99b496534348c1280ad909e760336dc4b8871a2490306683034038043d781b2ac482123faee320c9a074cd8763885628363c01bc809aabc4c84eba430e2b858a0c721fceb09c00acd253432961ccc47b4887c0744d05b1864034a32eaaba9b8590144a713c3e0e0d02971bfed4d94ea209d8c16e28000f184ab78508a43c602722d2d80a46118168c",
        "difficulty": "0x0",
        "number": "0x126b8c2",
        "gasLimit": "0x1c9c380",
        "gasUsed": "0xe717d7",
        "timestamp": "0x65dd187f",
        "extraData": "0x546974616e2028746974616e6275696c6465722e78797a29",
        "mixHash": "0x994f3286e689f095352b96c19f2ec710401f2a0f4936487de57a7c9f11b2a200",
        "nonce": "0x0000000000000000",
        "baseFeePerGas": "0x843fc6ac5",
        "withdrawalsRoot": "0x42107a0f6de27aa02963c41aca74f3683751424bd4cff7ccbb2e552b383cf7c7",
        "blobGasUsed": null,
        "excessBlobGas": null,
        "parentBeaconBlockRoot": null,
        "hash": "0xb493cb5d4cb5342f93783415cdb4700963af8ca1d4455d8af2499dbdb0ea7b46"
    },
    "logs": [
        {
            "address": "0x000000000022d473030f116ddee9f6b43ac78ba3",
            "topics": [
                "0xc6a377bfc4eb120024a8ac08eef205be16b817020812c73223e81d1bdb9708ec",
                "0x000000000000000000000000cb1ada11b21fe066dcb91a12cb8195fafa50420b",
                "0x0000000000000000000000003419875b4d3bca7f3fdda2db7a476a79fd31b4fe",
                "0x00000000000000000000000080a64c6d7f12c47b7c66c5b4e20e72bc1fcd5d9e"
            ],
            "data": "0x000000000000000000000000000000000000000000008dcbad0d69ce323e3e000000000000000000000000000000000000000000000000000000000066049e6e0000000000000000000000000000000000000000000000000000000000000000",
            "blockNumber": "0x126b8c2",
            "transactionHash": "0x37c21522d65111cfa101f99e4993f1c3ef02f9c70985bb90d3588de489c9f251",
            "transactionIndex": "0x0",
            "blockHash": "0xb493cb5d4cb5342f93783415cdb4700963af8ca1d4455d8af2499dbdb0ea7b46",
            "logIndex": "0x0",
            "removed": false
        },
    [...]

and the `/metrics` endpoint for a prometheus metrics endpoint:

    % curl -D - http://localhost:8989/metrics
    HTTP/1.1 200 OK
    Content-Type: text/plain; version=0.0.4; charset=utf-8
    Date: Mon, 26 Feb 2024 23:04:28 GMT
    Content-Length: 268
    
    # HELP blockchain_block_count number of blocks processed
    # TYPE blockchain_block_count counter
    blockchain_block_count 17
    # HELP blockchain_transaction_count number of transactions processed
    # TYPE blockchain_transaction_count counter
    blockchain_transaction_count 1808


The full set of arguments is:

* `--node`: websocket URL for the node.
* `--http-serve`: Bind address for the HTTP server. Default is `0.0.0.0:8989`.
* `--timeout`: I/O timeout in go-units. Default `15s`.

## Explanation

The monitor it doesn't do much useful. I took it more as a learning opportunity
but didn't have the time for much more.

It will subscribe to _newHead_ and _logs_ subscription feeds, and will correlate
them assuming that neither are interleaved. That is, logs are reported sequentially
without intermixing logs for different blocks.

It will aggregate each head(header) with it's associated blocks and then pass
this through a set of consumer that try to extract some basic metrics.

## Metrics

Just some basic metrics are exported:

- Number of transactions
- Number of new blocks produced.

Didn't have time to research for more security-oriented metrics so used the ones
avobe as a small example.

## TODO

Ideally I would have liked to be able to make sense of the logs / contracts,
for example by using the information in [this repo.](https://github.com/otterscan/topic0/tree/main/with_parameter_names)

Also the way the "metrics" are exposed, via prometheus endpoint, is possibly
not useful for a security use case, but it was simpler to implement.
