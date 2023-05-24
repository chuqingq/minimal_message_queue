## CA

#### 生成CA私钥

```openssl genrsa -out ca.key 2048```

#### 生成CA证书请求文件

```openssl req -new -key ca.key -out ca.csr -subj "/C=CN/ST=BJ/L=beijing/O=myorganization/OU=mygroup/CN=myname"```

#### 生成CA证书

```openssl x509 -req -days 365 -sha256 -extensions v3_ca -signkey ca.key -in ca.csr -out ca.cer```

## 服务端

#### 生成服务端私钥

```openssl genrsa -out server.key 2048```

#### 生成服务端证书请求文件

```openssl req -new -key server.key -out server.csr -subj "/C=CN/ST=BJ/L=beijing/O=myorganization/OU=mygroup/CN=myname"```

#### 生成服务端证书

```openssl x509 -req -days 365 -sha256 -extensions v3_req -CA ca.cer -CAkey ca.key -CAcreateserial -in server.csr -out server.cer```

## 客户端

#### 生成客户端私钥

```openssl genrsa -out client.key 2048```

#### 生成客户端证书请求文件

```openssl req -new -key client.key -out client.csr -subj "/C=CN/ST=BJ/L=beijing/O=myorganization/OU=mygroup/CN=myname"```

#### 生成客户端证书

```openssl x509 -req -days 365 -sha256 -extensions v3_req -CA ca.cer -CAkey ca.key -CAcreateserial -in client.csr -out client.cer```

