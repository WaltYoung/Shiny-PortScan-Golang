# 基于Golang实现的端口扫描器

## 功能

- 支持TCP连接扫描
- 支持SYN半开扫描
- 支持对IP/域名/CIDR网段进行扫描
- 支持大网段并发扫描
- 支持指定端口范围扫描
- 目前仅支持Windodws
- 目前仅支持IPv4

## 使用方法

1. 下载代码到本地
2. 修改参数：在`global.go`文件中，修改`Iface`变量的值为你想要使用的网卡
   - Windows下使用`ipconfig`查看各个网络适配器,如下图红框中所示的网络适配器的名称。
     ![Windows入参](https://github.com/user-attachments/assets/e820c4be-17f4-488e-8d45-12874d7dfd3c)
3. 使用以下命令编译代码：
   ```bash
   go build -o port_scanner.exe main.go
   ```
4. 运行编译后的程序
