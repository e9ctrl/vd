# vd 
In the complex realm of device integration, aligning multiple software layers is essential. However, if the communication layer falls short, the work invested in the subsequent layers may not bring value to the end-user.

Recognizing this challenge, this project was conceived to streamline the testing of communication layer and accelerate the process. Its design prioritizes simplicity in device testing, paving the way for effortless communication evaluations, particularly in CI/CD environments.

![](readme.gif)

# For who?
- Control system engineers who need to integrate a device:
    - That they don't have access to.
    - Whose access is restricted.
    - Accompanied by outdated documentation.
    - Where sending command data might be hazardous.

- Device suppliers keen on seamless integration of their device into a client's system.
- Security experts wanting to check if the IOC genuinely sends what it's supposed to.
- Everyone aiming to enhance the quality of their integrations.

# How It Works
The `vd` tool can simulate devices that communicate using stream-based communication, i.e., those that can be integrated using StreamDevice. Creating a simulator doesn't require programming; communication is described in a file called `vdfile`.

`vd` is able to lex and parse incoming messages from the client and, based on them, determine action.

# vdfile 
The configuration file called `vdfile` is a [TOML](https://toml.io/en/) file and captures communication details. 

Note that some parameters do not contain specific communication patterns. Your configuration can include one or multiple parameters depending on your needs.

It starts with information about terminators
```toml
[terminators]
intterm = "CR LF"
outterm = "CR LF"
```
and describes parameters available in the simulated device:

```toml
[[parameter]]
name = "current"
type = "int"
req = "CUR?"
res = "CUR %d"
set = "CUR %d"
ack = "OK"
val = 300
```

Here's a breakdown of the configuration:

* `name`: Parameter's name, not used in client communication but utilized in the HTTP API.
* `type`: Parameter type (available values - `int`, `float`, `string`, `bool`).
* `req`: Client's request to the simulated device to read the value.
* `res`: The response the simulated device sends to the client for the request.
* `set`: Client's request to change the parameter value.
* `ack`: Device's acknowledgment to the set request.
* `val`: Contains the default value. Note that for int and float, the value should be without quotes.
* `opt`: (Optional) Limits the range of values a parameter can take (see below for example of usage).

Below is a sample configuration:
```toml
[[parameter]]
name = "version"
type = "string"
req = "ver?"
res = "%s"
val = "version 1.0"

[[parameter]]
name = "current"
type = "int"
req = "CUR?"
res = "CUR %d"
set = "CUR %d"
ack = "OK"
val = 300

[[parameter]]
name = "temperature"
type = "float"
req = "TEMP?"
res = "%.2f"
set = "TEMP %.2f"
ack = "%.2f"
val = 36.6

[[parameter]]
name = "mode"
type = "string"
opt = "NORM|SING|BURS|DCYC"
req = "MODE?"
res = "%s"
set = "MODE %s"
ack = "ok"
val = "NORM"
```
# Installation
`vd` is supplied as a binary file. Download the appropriate version for your operating system and you are good to go.

# Usage
After creating the configuration file, you can launch the program with:

```bash
$ vd <path to vdfile>
```

By default, the simulator is accessible on port `:9999`. You can connect to this port using an IOC or via telnet.

The simulator also starts an HTTP server with an API that allows direct parameter value changes via HTTP. By default, it listens on port `:8080`.

To fetch the current value of a parameter, e.g., temperature:
```bash
$ curl localhost:8080/temperature
```
To set a new value:
```bash
$ curl -X POST localhost:8080/temperature/36.6
```

To make our lifes easier, `vd` comes with a built-in client. You can use it as follows:
```
$ vd get temperature
$ vd set temperature 36.6
```


If in doubt, check the help
```
$ vd -h
```
