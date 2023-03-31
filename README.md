> **DEPRECATION NOTICE:** We've stopped using Zammand and hence don't need this
> integration anymore. While it might still work perfectly fine, we currently
> don't have the capacity to maintain it anymore.

Asterisk integration for Zammad
===============================

This works for me.
If you need support, just [contact me](https://www.digineo.de/impressum).

## Installation

    go install github.com/digineo/zammad-asterisk@latest

## Configuration

### Asterisk

Create a new account for the [Asterisk REST Interface](https://wiki.asterisk.org/wiki/pages/viewpage.action?pageId=29395573) by editing the file `ari.conf` in the asterisk configuration directory:

```ini
[general]
enabled = yes

[zammad]
type = user
read_only = no
password = secret5
```

Add the application to your Dialplan in a context for incoming calls.
The second argument for `Stasis()` is the name of called destination.
If you have several numbers for incoming calls you can use this argument to distinguish between them.

```
context incoming {
	12345678 => {
		Stasis(zammad, foobar);
		Dial(...);
		Hangup;
	}
}
```

### Zammad Interface

Create a `config.cfg` with the following configuration:

```ini
[asterisk]
host     = "127.0.0.1"
port     = 8088
username = "zammad"
password = "secret5"

[zammad]
endpoint = "https://zammad.example.com/api/v1/vti_logs"
token    = "your secret token"
```

## Running

    zammad-asterisk path/to/config.cfg
