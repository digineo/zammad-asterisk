Asterisk integration for Zammad
===============================

This is **work in progress**!

## Installation

    go get github.com/digineo/zammad-asterisk/...

## Configuration

### Asterisk

Create a new account for the [Asterisk Manager Interface](http://the-asterisk-book.com/1.6/asterisk-manager-api.html) by creating the file `manager.d/zammad.conf` in the asterisk configuration directory:

```
[zammad]
secret = secret5
deny = 0.0.0.0/0.0.0.0
permit = 127.0.0.1/255.255.255.255
read = call
```

### Zammad Interface

Create a `config.cfg` with the following configuration:

```
[asterisk]
endpoint = "127.0.0.1:5038"
username = "zammad"
password = "secret5"

incoming = [
  "SIP/your_incoming_channel",
]

[zammad]
endpoint = "https://zammad.example.com/api/v1/asterisk/in"
```

## Running

    zammad-asterisk path/to/config.cfg
