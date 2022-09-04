# 描述

将文件（图片、视频、文本等）上传到自建minio服务器的工具。

该工具是《使用minio搭建自己的图床》的一部分，可以通过该工具方便的实现图片上传功能。

# 使用方法

## 编译

```shell
bash ./build.sh
```

## 安装

```bash
cp ./release/light_minio_client.linux.amd64 /usr/bin/light_minio_client
```

## 配置

linux下配置文件：`$HOME/.light-minio-client.json`

windows下配置文件：`%HOMEPATH%/.light-minio-client.json`

MacOS下配置文件：`$HOME/.light-minio-client.json`

```json
{
  "endpoint": "minio服务器地址",
  "access_key_id": "xxx",
  "secret_access_key": "xxxx",
  "bucket_name": "桶名称",
  "is_ssl": true,
  "content_types": {
    ".jpg": "image/jpg",
    ".png": "image/png",
    ".gif": "image/gif",
    ".bmp": "image/bmp",
    ".txt": "text/plain",
    ".log": "text/plain"
  }
}
```



## 运行

```bash
/usr/bin/light_minio_client test.txt test.jpg
```

每一行对应一个参数的上传结果。如果上传成功，是文件的url地址；如果失败，是错误信息。

# Typora中自动上传图片

![image-20220904221606718](https://home.mydata.top:8684/blog/20220904221606-image-20220904221606718.png)