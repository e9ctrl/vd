[![Latest Release](https://img.shields.io/github/release/e9ctrl/vd.svg?style=flat-square)](https://github.com/e9ctrl/vd/releases/latest)
[![Tests](https://github.com/e9ctrl/vd/actions/workflows/tests.yml/badge.svg)](https://github.com/e9ctrl/vd/actions/workflows/tests.yml)
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Report](https://goreportcard.com/badge/github.com/e9ctrl/vd)](https://goreportcard.com/report/github.com/e9ctrl/vd)
![coverage](https://raw.githubusercontent.com/e9ctrl/vd/badges/.badges/main/coverage.svg)

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

It starts with optional error message:
```toml```
mismatch = "error"
```
then, it contains information about terminators:
```toml
[terminators]
  intterm = "CR LF"
  outterm = "CR LF"
```
and finally describes parameters available in the simulated device:

```toml
[[parameter]]
  name = "current"
  typ = "int64"
  val = 300

[[command]]
  name = "get_current"
  req = "CUR?"
  res = "CUR {%d:current}"

[[command]]
  name = "set_current"
  req = "CUR {%d:current}"
  res = "OK"
  dly = "1s"

[[parameter]]
  name = "state"
  typ = "string"
  val = "on"
  opt = "ON|OFF|ERROR"

[[command]]
  name = "set_state"
  req = "SET {%s:state}"
  res = "OK"
```

Here's a breakdown of the configuration:

* `name`: Parameter's name, not used in client communication but utilized in the HTTP API.
* `typ`:  Parameter type (available values - `int`, `float`, `string`, `bool`).
* `req`:  Client's request to the sumylated device to get or set value.
* `res`:  The response the simulated device sends to the client for the request.
* `dly`:  Response delay with time unit.
* `opt`: (Optional) Limits the range of values a parameter can take (see below for example of usage).


Below is a sample configuration:
```toml

mismatch = "Wrong query"

[[terminators]]
  intterm = "CR LF"
  outterm = "CR LF"

[[parameter]]
  name = "version"
  typ = "string"
  val = "version 1.0"

[[command]]
  name = "get_version"
  req = "ver?"
  res = "{%s:version}"
  dly = "10s"

[[parameter]]
  name = "current"
  typ = "int64"
  val = 300

[[command]]
  name = "get_current"
  req = "CUR?"
  res = "CUR {%d:current}"

[[command]]
  name = "set_current"
  req = "CUR {%d:current}"
  res = "OK"
  dly = "2s"

[[parameter]]
  name = "temperature"
  typ = "float64"
  val = 36.6

[[command]]
  name = "get_temp"
  req = "TEMP?"
  res = "{%.2f:temperature}"

[[command]]
  name = "set_temp"
  req = "TEMP {%.2f:temperature}"
  res = "{%.2f:temperature}"

[[parameter]]
  name = "mode"
  typ = "string"
  opt = "NORM|SING|BURS|DCYC"
  val = "NORM"

[[command]]
  name = "get_mode"
  req = "MODE?"
  res = "{%s:mode}"

[[command]]
  name = "set_mode"
  req = "MODE {%s:mode}"
  res = "ok"

[[command]]
  name = "get_status"
  req = "status?"
  res = mode:{%s:mode},temp:{%2f:temperature}

```

# Parameter
`parameter` is a place where parameter together with its name, type, possible values, and initial value are defined. 

# Command
`command` is section that keeps information about accepted request strings and responses to them. The command can reference none, one or more parameters. One can assign command to the parameter using `{` `}` with proper placeholder and parameter name between brackets e.g. `{%d:parameter}`.

# Delays
The `vd` tool enables the introduction of delays when sending responses to requests. This feature allows you to define custom wait times for the `vd` to hold off on every response and acknowledgment, enhancing the simulation of real-world network conditions or server response times.

The delays are specific for single command, you can use `dly` to define delay time for the given command.

# Mismatch
`vd` allows to specify mismatch that is sent back to the client when received string does not match any of the expected commands. It is send back to the client automatically without delay.

# Triggering reply
The `vd` tool enables the triggering of responses, simulating scenarios where a device sends data autonomously, without a specific request from the client. It is done by sending proper request via HTTP API. 

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

To make our lifes easier, `vd` comes with a built-in client.

To change the value of a parameter, e.g., temperature:
```
$ vd get temperature
$ vd set temperature 36.6
```

To change the value of command delay:
```
$ vd get delay get_status
$ vd set delay get_status 200ms
```

To change mismatch message string:
```
$ vd get mismatch
$ vd set mismatch "wrong message"
```

To trigger a response related to a parameter, e.g., temperature:
```
$ vd trigger temperature
```

If in doubt, check the help
```
$ vd -h
```
