## Goplum IRC notifier plugin

Plugin for [IRC-Bot](https://github.com/greboid/irc-bot)

Receives notifications from a [goplum](https://github.com/csmith/goplum) instance and outputs them to a channel.

 - go build go build github.com/greboid/irc-goplum/v2/cmd/goplum
 - docker run greboid/irc-goplum
 
#### Configuration

At a bare minimum you also need to give it a channel, a secret to use as part of the URL to receive notifications
 on and an RPC token.  You'll like also want to specify the bot host.

Once configured the URL to configure in goplum would be <Bot URL>/goplum/<secret>

#### Example running

```
---
version: "3.5"
service:
  goplum:
    image: greboid/irc-goplum
    environment:
      RPC_HOST: bot
      RPC_TOKEN: <as configured on the bot>
      CHANNEL: #spam
      SECRET: cUCrb7HJ
```

```
goplum -rpc-host bot -rpc-token <as configured on the bot> -channel #spam -secret cUCrb7HJ
```
