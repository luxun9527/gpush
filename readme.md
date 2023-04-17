# 轻量的ws推送系统

## 简介

两个部分组成

![image-20220404210048689](C:\Users\deng\AppData\Roaming\Typora\typora-user-images\image-20220404210048689.png)

proxy 面向后台服务

api 面向客户端，通过grpc (stream)连接。

## api

重点介绍api
todo 一般是写多读少的情况使用epoll来避免没一个连接都创建goroutine 从tcp upgrade 避免tcp 4k 内存的分配

1、推送方式

​	1.1 推送对个人

​	1.2 推送给所有

​	1.3 推送给指定的topic

​	1.4 消息合并（消息达到一定数量，或者到时间推送一次）

2、连接管理

​	2.1每个topic限制订阅的人数

​	2.2限流，每秒钟最多连接多少人次

​	2.3 反复断开重连的拉黑

3、统计

​	3.1统计在线连接数

​	3.2统计推送的情况

4、客户端

  4.1订阅

 4.1取消订阅

