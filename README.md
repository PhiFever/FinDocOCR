## 参考文档

### 百度云

智能财务票据识别1个月只有200次……

- 错误码：https://cloud.baidu.com/doc/IMAGERECOGNITION/s/Lk3bcxeoc
- 控制台：https://console.bce.baidu.com/ai/?_=1706703162407#/ai/speech/app/list
- 鉴权认证（Access Token）：https://ai.baidu.com/ai-doc/REFERENCE/Ck3dwjhhu
- 领取免费测试额度：https://console.bce.baidu.com/ai/#/ai/ocr/overview/resource/getFree
- 智能财务票据识别：https://cloud.baidu.com/doc/OCR/s/7ktb8md0j

### 腾讯云

通用票据识别（高级版）每月免费100次，QPS能拉到5

- [票据单据识别-产品简介](https://cloud.tencent.com/document/product/866/37495)
- [腾讯云 - 控制台](https://console.cloud.tencent.com/ocr/overview)
- [通用票据识别（高级版）API 文档](https://cloud.tencent.com/document/product/866/90802)
  - [OCR Demo](https://ocrdemo.cloud.tencent.com/)

## 阿里云

- [票据凭证(OCR)-产品简介](https://help.aliyun.com/zh/ocr/product-overview/ticket-and-invoice-recognition-1)

## 使用方法

1. 在百度云控制台创建应用，获取`ClientId`和`ClientSecret`。相关步骤可以[
   参考](https://blog.csdn.net/m0_49710816/article/details/122743651)
2. 在同级目录下创建`.env`文件，填入以下内容：

```toml
ClientId = "你自己的"
ClientSecret = "你自己的"
DOC_DIR="docs"
```

3. 在`$DOC_DIR`目录下放入需要识别的图片及单页pdf（本程序暂时不支持多页pdf，虽然百度云支持）。
4. 运行`main.go`，等待程序自动识别图片并输出结果到项目根目录目录的.xlsx文件中。

## TODO
1. ~~使用[bild](https://github.com/anthonynsimon/bild)替换很久没有维护的imaging库~~
2. ~~使用策略模式重构`main.go`中的保存结果部分代码~~
3. 添加错误日志邮件压缩发送