# Scheduler
distributed task scheduling by go

[reference](https://github.com/imlgw/scheduler)


## 分布式架构的核心 要素

1. 调度器，**高可用**,由于分布式部署，确保不会由于单点故障停止调度，同时任务数据也是保存在etcd中的，不容易的丢失
2. 执行器，能够具有横向扩展能力,能够提供大量任务的并行处理能力

## CAP理论（常用于分布式存储）

![image-20220223210834510](https://gitee.com/Euraxluo/images/raw/master/picgo/image-20220223210834510.png)

C：一致性，写入后立即能读取到新值

A：可用性，通常保障最终一致性，因为高可用性必然无法保证一致性

P：分区容错性，必须保证，因为分布式一定需要面对网络分区

## BASE 理论（常用于应用架构）

![image-20220223211323437](https://gitee.com/Euraxluo/images/raw/master/picgo/image-20220223211323437.png)

BA（基本可用）：损失部分可用性，保证整体可用性，例如熔断机制等

S（软状态）：允许状态同步延迟，只要不影响系统即可

E（最终一致性）：经过一段时间后，系统能够达到一致性就好

## Master 架构
![master.png](master.png)

## Worker 架构
![worker.png](worker.png)


## issue
kill 时产生孤儿进程 导致无法立刻回收协程的问题

