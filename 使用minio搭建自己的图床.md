图床是啥？就是一个专门用来存储照片、视频的仓库。因为我们在写blog、markdown的时候都会用到各种照片或者视频。通常我们会把他们放在文档一起的地方。比如blog网站里面。但是，如果你有多个blog，或者日后需要迁移blog的时候就会比较烦。所以，就会把图片专门放在一个网站中管理，也就是图床。在网上有各种免费、收费的图床网站。免费的缺点很明显，服务会差一些。收费的也难说哪天就消失了。于是，想着我既然有服务器，何不自己搭建一套图床出来呢。

# minio

minio是一套开源的对象存储服务器。它提供了内容存储功能，包括图片、视频和其他各种类型的文件。关键是它非常轻量级，而且还提供了docker部署方式，有提供了如java、Python、GO等语言的SDK。所以，我们可以：

1. 很方便的将它部署起来；
2. 可以很方便地使用各种语言写客户端调用它，实现上传、下载等等功能。
3. 可以与各种工具进行集成，如picgo、typora等等。

## 部署minio

下载很简单，一行命令即可：

```shell
docker pull minio/minio
```

运行也简单，只需要使用`docker run`即可。但是，为了方便管理，我还是写了一个很简单脚本——`startup.sh`

```bash
#!/bin/bash
name=minio
user_name=test
user_password=test
console_port=9011
data_port=9010
data_path=/home/minio

docker rm -f $name

param="-d --restart=always --name $name"
param="${param} -v $data_path/config:/root/.minio"
param="${param} -v $data_path/data:/data"
param="${param} -v /etc/localtime:/etc/localtime:ro"
param="${param} -v /etc/timezone:/etc/timezone:ro"
param="${param} -e MINIO_ACCESS_KEY=$user_name"
param="${param} -e MINIO_SECRET_KEY=$user_password"
param="${param} -p $console_port:$console_port -p $data_port:$data_port"

docker run ${param} \
     minio/minio server \
     /data --console-address ":$console_port" -address ":$data_port"
```

接下来，运行一下`bash startup.sh`即可启动minio服务了。通过浏览器打开`http://127.0.0.1:9011`就可以打开minio的控制台了。

## 创建bucket

bucket，也就是存储桶。你可以理解成分组，以便把相同用途的文件分类存储。例如博客用到的图片是可以公开的，所以，专门建一个可以公开访问的桶，而那些私有的，不允许公开访问的文件放在私有桶中。又或者不同的项目也可以放在不同的桶中。这儿我们创建一个可以公共访问的桶`blog`

1. 创建桶

![image-20220903232805386](https://home.mydata.top:8684/blog/20220903232813.png)

![image-20220903232904916](https://home.mydata.top:8684/blog/20220903232916.png)

2. 将桶的权限设置为公开

   ![image-20220903233311208](https://home.mydata.top:8684/blog/20220903233314.png)

![image-20220903233344305](https://home.mydata.top:8684/blog/20220903233347.png)

注意：如果不将访问权限设置为公开，会无法直接访问。

当前，还可以更加保守一些，将起设置为只读。

![image-20220903233524292](https://home.mydata.top:8684/blog/20220903233528.png)

## 创建用户【可选】

为了安全，一般不会直接使用管理员账号来上传和下载图片，而是新建一个专门的账号。这样一旦账号信息泄露，可以随时改密码或者删除掉它。同时，也可以给这个账户有限的权限，以尽量降低信息泄露带来的安全风险。当然，minio还提供了更安全的accessKey和secretKey。通过key访问服务器，即使泄露了信息，也可以方便的调整。

创建用户的方法很简单，如下：

![image-20220903234018670](https://home.mydata.top:8684/blog/20220903234022.png)

![image-20220903234117226](https://home.mydata.top:8684/blog/20220903234120.png)

注意：权限这个地方，一定不要选择管理权限。对于公共桶，建议选择读写或者只写权限。对于私有桶，建议设置两个账号，一个只读、一个只写或者读写。权限不是越大越好，而是够用就好。权限越大，风险越高，一旦账号出现问题带来的危害也就越大。

# 上传文件到minio

minio提供了多种上传接口，可以在consol的网页中上传，可以用python写脚本上传，也可以通过S3协议上传…

## 网页上传

网页上传需要使用具有consoleAdmin权限的账号，比如我们启动docker时指定的账号，就有该权限。当然，也可以是后建的拥有consoleAdmin权限的账号。

首先，选择目标桶，点击`Browse`浏览该桶的内容。

![image-20220903234715950](https://home.mydata.top:8684/blog/20220903234719.png)

然后，点击`Upload`上传文件到目标桶。

![image-20220903234750437](https://home.mydata.top:8684/blog/20220903234754.png)

## 使用python脚本上传

minio提供了python语言的sdk，我们可以很方便的使用该SDK上传文件到minio。简单的脚本`minio-client.py`如下：

```python
import os
import time
import uuid
import sys
import requests
from minio import Minio
from minio.error import MinioException
import warnings

ip = "127.0.0.1"
port = "9010"
accessKey = "accessKey"
secretKey = "secretKey"
isSSl = False
bucket = "blog"
protcol = "http://"
if isSSl:
    protcol = "https://"

warnings.filterwarnings('ignore')
files = sys.argv[1:]
minioClient = Minio(ip+":"+port,
                    access_key=accessKey, secret_key=secretKey, secure=isSSl)
result = ""
date = time.strftime("%Y%m%d%H%M%S", time.localtime())

for file in files:
    file_name = os.path.split(file)[-1]
    file_type = os.path.splitext(file)[-1]
    new_file_name = "{}-{}".format(date, file_name)
    file_type = os.path.splitext(file_name)[-1]
    if file_type in [".png", ".jpg", ".gif"]:
        content_type = "image/" + file_type.replace(".", "")
    elif file_type in [".py", ".txt", ".log"]:
        content_type = "text/plain"
    else:
        result = result + "error: uploda {} failed, not support file type\n".format(file)
        continue
    try:
        minioClient.fput_object(bucket_name=bucket, object_name=new_file_name, file_path=file, content_type=content_type)
        result = result + protcol + ip + ":" + port + "/" + bucket + "/"  + new_file_name + "\n"
    except MinioException as err:
        result = result + "error:" + err.message + "\n"
print(result)
```

注意：

1. 脚本中，我们仅仅支持了少数几种类型的数据，您可以根据需要支持更多的类型。

2. 其中的`accessKey`和`secretKey`可以是用户名和密码，也可以是minio提供的`Service Account`。为了安全，建议使用Service Account，创建方法如下：

   ![image-20220904001820881](https://home.mydata.top:8684/blog/20220904001851-image-20220904001820881.png)

![image-20220904001928608](https://home.mydata.top:8684/blog/20220904001932-image-20220904001928608.png)

![image-20220904002020695](https://home.mydata.top:8684/blog/20220904002025-image-20220904002020695.png)

完成后会下载到一个json文件，其中就有`accessKey`和`secretKey`。

在运行脚本前，还需要安装一下用到的库：

```bash
pip3 install minio requests
```

安装完成后就可以正常使用了，使用方法：

```bash
python3 minio-client.py test.jpg test.png
```

不出意外的话就可以上传成功了，例如：

![image-20220904002704339](https://home.mydata.top:8684/blog/20220904002704-image-20220904002704339.png)

## 使用go语言客户端上传

[https://github.com/xxNull-lsk/light-minio-client](https://github.com/xxNull-lsk/light-minio-client)

# typora中自动上传图片

非常简单，按照如下设置即可。设置完成后typora中插入的图片就可以自动上传到我们自建的图床中了。

![image-20220904002806835](https://home.mydata.top:8684/blog/20220904002807-image-20220904002806835.png)

# 互联网访问自建图床

如果图床只能在本机访问，其实意义不大。通常，我们需要可以通过互联网随时随地访问它。那么，你就需要一个服务器。可以是公有云，如阿里云、腾讯云等，提供的ECS主机或者应用主机。通常他们，会提供一个公网IP。你把minio部署到这些服务器上就可以了。

如果，你有办法搞到自家的公网IP，当然，也可以把服务部署自己家里的服务器上。然后，再通过家里的公网IP或者域名访问。至于如何搞到自家的公网IP，就需要一点点技巧了。在网上可以查到相关的资料。比如百度搜素：联通  公网IP 等关键字。

# 添加水印

当然，还可以在上传之前自动添加水印，完整代码如下：

```python
from PIL import Image, ImageDraw, ImageFont, ImageEnhance, ImageChops
import math
import os
import time
import uuid
import sys
import requests
from minio import Minio
from minio.error import MinioException
import warnings


ip = "127.0.0.1"
port = "9010"
accessKey = "accessKey"
secretKey = "secretKey"
isSSl = False
bucket = "blog"
protcol = "http://"
if isSSl:
    protcol = "https://"

mask1 = {"opacity": 0.02,
         "size": 32,
         "space": 60,
         "angle": 30,
         "color": "#F4EEF1",
         "mark": "https://blog.mydata.top",
         "font": 'NotoMono-Regular.ttf'}

mask2 = {"opacity": 0.1,
         "size": 64,
         "color": "#00FF00",
         "mark": "如斯说",
         "font": 'DroidSansFallbackFull.ttf'}


# 裁剪图片边缘空白
def crop_image(im):
    bg = Image.new(mode='RGBA', size=im.size)
    diff = ImageChops.difference(im, bg)
    del bg
    bbox = diff.getbbox()
    if bbox:
        return im.crop(bbox)
    return im


# 设置水印透明度
def set_opacity(im, opacity):
    assert opacity >= 0 and opacity <= 1

    alpha = im.split()[3]
    alpha = ImageEnhance.Brightness(alpha).enhance(opacity)
    im.putalpha(alpha)
    return im


# 在im图片上添加水印 im为打开的原图
def mark_im_fill(im, mark):

    # 计算斜边长度
    c = int(math.sqrt(im.size[0] * im.size[0] + im.size[1] * im.size[1]))

    # 以斜边长度为宽高创建大图（旋转后大图才足以覆盖原图）
    mark2 = Image.new(mode='RGBA', size=(c, c))

    # 在大图上生成水印文字，此处mark为上面生成的水印图片
    y, idx = 0, 0
    while y < c:
        # 制造x坐标错位
        x = -int((mark.size[0] + args['space']) * 0.5 * idx)
        idx = (idx + 1) % 2

        while x < c:
            # 在该位置粘贴mark水印图片
            mark2.paste(mark, (x, y))
            x = x + mark.size[0] + args['space']
        y = y + mark.size[1] + args['space']

    # 将大图旋转一定角度
    mark2 = mark2.rotate(args['angle'])

    # 在原图上添加大图水印
    if im.mode != 'RGBA':
        im = im.convert('RGBA')
    im.paste(mark2,  # 大图
             (int((im.size[0] - c) / 2), int((im.size[1] - c) / 2)),  # 坐标
             mask=mark2.split()[3])
    del mark2
    return im


# 在im图片上添加水印 im为打开的原图
def mark_im_br(im, mark):
    # 在原图上添加大图水印
    if im.mode != 'RGBA':
        im = im.convert('RGBA')
    im.paste(mark,  # 大图
             (int((im.size[0] - mark.size[0]) - 15),
              int((im.size[1] - mark.size[1]) - 15)),  # 坐标
             mask=mark.split()[3])
    return im


# 生成水印图片，返回添加水印的函数
def generate_watermark_image(args):
    # 字体宽度
    width = len(args["mark"]) * args["size"]

    # 创建水印图片(宽度、高度)
    mark = Image.new(mode='RGBA', size=(width, args['size'] + 8))

    # 生成文字
    draw_table = ImageDraw.Draw(im=mark)
    draw_table.text(xy=(0, 0),
                    text=args["mark"],
                    fill=args['color'],
                    font=ImageFont.truetype(args['font'], size=args['size'])
                    )
    del draw_table

    # 裁剪空白
    mark = crop_image(mark)

    # 透明度
    set_opacity(mark, args['opacity'])
    return mark


def mark_image(args, fn_mark, im):
    # im = Image.open(args['file'])
    mask = generate_watermark_image(args)
    im = fn_mark(im, mask)
    return im


warnings.filterwarnings('ignore')
files = sys.argv[1:]
minioClient = Minio(ip+":"+port,
                    access_key=accessKey, secret_key=secretKey, secure=isSSl)
result = ""
date = time.strftime("%Y%m%d%H%M%S", time.localtime())

for file in files:
    file_name = os.path.split(file)[-1]
    file_type = os.path.splitext(file)[-1]
    new_file_name = "{}-{}".format(date, file_name)
    file_type = os.path.splitext(file_name)[-1]
    if file_type in [".png", ".jpg", ".gif"]:
        content_type = "image/" + file_type.replace(".", "")
        im = Image.open(file)
        im = mark_image(mask1, mark_im_fill, im)
        im = mark_image(mask2, mark_im_br, im)
        file = "/tmp/{}".format(new_file_name)
        # im.save(file)
    elif file_type in [".py", ".txt", ".log"]:
        content_type = "text/plain"
    else:
        result = result + \
            "error: uploda {} failed, not support file type\n".format(file)
        continue
    try:
        minioClient.fput_object(
            bucket_name=bucket, object_name=new_file_name, file_path=file, content_type=content_type)
        result = result + protcol + ip + ":" + port + \
            "/" + bucket + "/" + new_file_name + "\n"
    except MinioException as err:
        result = result + "error:" + err.message + "\n"
    if file.startswith('/tmp/'):
        os.remove(file)
print(result)

```

效果如下：

![image-20220904015708516](https://home.mydata.top:8684/blog/20220904015708-image-20220904015708516.png)
