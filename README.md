![Odometer](./odometer.png)

> Monitor a Gas or Water Odometer with a Raspberry Pi

## What is this?

It is a small program, written in Go that is designed to run on a Raspberry Pi.

It starts a Prometheus server on port 8080.

The [GPIO Pin 26](https://pinout.xyz/pinout/pin37_gpio26) should be connected
to a Reed switch.

![circuit sketch](./sketch.jpg)

## Deploy To Balena

```shell
git clone https://github.com/stayradiated/pi-odometer
balena push odometer
```
