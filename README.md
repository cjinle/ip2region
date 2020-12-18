# ip2region根据IP查询地区的http服务



## docker 使用
```sh
sudo docker pull chenjinle/ip2region:0.0.2

sudo docker run --name ip2region2 -d -p "8080:8080" chenjinle/ip2region:0.0.2
```

## 测试

```
curl http://192.168.56.101:8080/8.8.8.8

# 返回 {"cityid":166,"country":"美国","region":"0","province":"0","city":"0","isp":"Level3"}
```


