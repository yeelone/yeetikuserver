
![image](https://wx4.sinaimg.cn/mw1024/6547935dgy1flmlznndy4j202y03cq2s.jpg)

# [忆题库 YeeTiku](http://yeetiku.com/) - 可定制的题库平台


YeeTiku分为三个部分：

1. [yeetiku server](https://github.com/yeelone/yeetiku-server-go) : golang+postgresql实现的服务器程序
2. [yeetiku admin](https://github.com/yeelone/yeetiku-admin) :react +dva实现的后台管理系统
3. [yeetiku mobile](https://github.com/yeelone/yeetiku-mobile-rn)：react native 实现的移动APP

react native 发布在expo上，有兴趣的朋友可以前往体验：
https://expo.io/@yeelone/yeetiku
![image](https://wx3.sinaimg.cn/mw1024/6547935dgy1flmm1qu1nnj20e1062t8o.jpg)


APP截图：
![image](https://wx4.sinaimg.cn/mw690/6547935dgy1flmmcbufv8j20fq0drwfg.jpg)
![image](https://wx1.sinaimg.cn/mw1024/6547935dgy1flmmcbu30ij20fi0dk0tu.jpg)

#### Basic usage

```
生成二进制服务器程序
G:\yeetiku\server-go> gox -os "linux" -arch amd64

root@DESKTOP-SRNKU4B:/mnt/g/yeetiku/server-go# ./server-go_linux_amd64 -h
Usage of ./server-go_linux_amd64:
  -config string
        specified config file
  -port int
        designated ports
  -reset-admin-password string
        reset admin password
root@DESKTOP-SRNKU4B:/mnt/g/yeetiku/server-go# ./server-go_linux_amd64 --port 8080
```

##### Next Step
- [ ] 后台题目支持图片
- [ ] 优化go代码，去除冗余
- [ ] 优化reactnative，还有一些bugs没处理好，比如登录界面跳转到主界面之后，返回仍然会回退到登录界面


##### Contributing

author: elone

email: yljckh@gmail.com
