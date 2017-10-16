# timer exampler

## flow
```flow
mylogic=>start: myLogic
timer=>inputoutput: Timer
polling=>operation: polling...
expire=>condition: expire?
handler=>end: Handler
mylogic->timer->polling->expire
expire(yes)->Handler
expire(no)->polling
```

- myLogic负责产生消息，Handler需要在N秒后处理消息。
- myLogic把消息推送给Timer，并附上消息目的地为Handler
- Timer计算出消息的超时时间戳，每次轮询都会检测自己维护的消息列表是否有消息过期，如果消息过期，则推送给对应的目的地。
- Handler处理收到的消息。

## RUN
```
cd handler
go run handler.go
cd timer_1s
go run timer_1s.go
cd mylogic
go run mylogic.go
```
