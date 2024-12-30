<p align="center">
  <a href="http://amywxd.site:3090">
    <img style="border-radius: 70px" src="https://raw.githubusercontent.com/BlingCc233/go_sudoku/refs/heads/main/assets/logo.png" width="300" height="300" alt="Cc-bot">
  </a>
</p>

<div align="center">

# Go_Sudoku

> > > > > 一种全新的基于socks5的低熵混淆代理协议


</div>

## 声明

> [!NOTE]\
> 仅供学习，请勿用于非法用途


  <br/>

## 启动

- 确保你有[Go](https://golang.org/)环境，版本需大于等于1.20
- 运行`go mod tidy`

### 客户端

- 在`cmd/sudosocks-local`下运行`go run main.go`
- 本地socks5端口默认为7789，远程地址端口默认为127.0.0.1:17789
- 通过参数`-l`和`-r`来修改本地和远程地址和端口

### 服务端

- 在`cmd/sudosocks-server`下运行`go run main.go`
- 运行端口默认为17789
- 通过参数`-p`来修改运行端口

## 功能

| 功能              | 说明           |
|-----------------|--------------|
| 数据基于4x4数独编码     | 能够做到低熵和混淆    |
| 基于自实现的协议头       | 实现了tls混淆     |
| 严格考虑了Wall的启发式规则 | 同时遵守了Ex1，Ex4 |
| 头部预留了混淆单元       | 防止主动探测       |

## 施工中的功能

- [ ] 日志分级
- [ ] socks在local侧处理
- [ ] 传输层协议自定义
- [ ] 一键部署脚本

## 鸣谢

- [链接1](https://gfw.report/publications/usenixsecurity23/zh/)
- [链接2](https://github.com/enfein/mieru/issues/8)
- [链接3](https://github.com/zhaohuabing/lightsocks)
- [链接4](https://imciel.com/2020/08/27/create-custom-tunnel/)
- [链接5](https://oeis.org/A109252)
- [链接6](https://pi.math.cornell.edu/~mec/Summer2009/Mahmood/Four.html)


