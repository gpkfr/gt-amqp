# gt-amqp
Minimalist Consumer and Publisher to RabbitMQ (aka AMQP) in Go

##Usage :

 ###Consumer creation syntax :

```shell
 ./gt-ampq -c config.ini -s production receive
```
 ### Sending message syntax :

  here we send  "-I http://google.com" to the queue gtq.

```shell
 $ ./gt-amqp -c config.ini -s production send -I http://google.com
```