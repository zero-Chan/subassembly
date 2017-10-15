# timer exampler

## flow
```flow
logica=>start: Logica
timer=>inputoutput: Timer
polling=>operation: polling...
expire=>condition: expire?
logicb=>end: Logicb
logica->timer->polling->expire
expire(yes)->logicb
expire(no)->polling
```

- Logica负责产生消息，Logicb需要在N秒后处理消息。
- Logica把消息推送给Timer，并附上消息目的地为Logicb
- Timer计算出消息的超时时间戳，每次轮询都会检测自己维护的消息列表是否有消息过期，如果消息过期，则推送给对应的目的地。
- Logicb处理收到的消息。
