## os-csi

为 k8s 动态使用对象存储设计的 csi 插件

⚠️ 未在生产环境严谨测试，仅供学习使用

#### 用法

参考 deploy/ 目录

部署流程

```bash
kubectl apply -f csi-provisioner.yaml
kubectl apply -f csi-os.yaml
```

修改配置，举例如下

`examples/secret.yaml`

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: csi-os-secret
  namespace: kube-system
stringData:
  accessKeyID: xxxx
  secretAccessKey: xxxx
  # endpoint 写访问地址，例如 <region>.xxx.xxx 不包含 bucket 名字
  endpoint: http://cos.ap-beijing.myqcloud.com
  # 使用的挂载命令，选择时注意镜像是否支持，默认可选 cosfs、ossfs。也用来区分不同服务商
  mounter: cosfs
```

`examples/storageclass.yaml`

```yaml
---
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: csi-os
provisioner: csi.os
parameters:
  # -d,--debug  set log level to debug 
  options: ""
  bucket: test-xxxx
  csi.storage.k8s.io/provisioner-secret-name: csi-os-secret
  csi.storage.k8s.io/provisioner-secret-namespace: kube-system
  csi.storage.k8s.io/controller-publish-secret-name: csi-os-secret
  csi.storage.k8s.io/controller-publish-secret-namespace: kube-system
  csi.storage.k8s.io/node-stage-secret-name: csi-os-secret
  csi.storage.k8s.io/node-stage-secret-namespace: kube-system
  csi.storage.k8s.io/node-publish-secret-name: csi-os-secret
  csi.storage.k8s.io/node-publish-secret-namespace: kube-system
```

应用配置

```bash
kubectl apply -f examples/secret.yaml
kubectl apply -f examples/storageclass.yaml
```

测试创建 pv，并挂载

```bash
kubectl apply -f examples/pvc.yaml
kubectl get pvc csi-os-pvc
kubectl apply -f examples/pod.yaml
# sleep 10 # wait for starting
kubectl get pv 
```


#### Todo

0. 完善用法说明

1. 解决启动顺序不同，没有正确在 ds pod 挂载的问题

2. 解决 ds pod 重启可能导致异常

3. 使用自定义 prefix 创建 pvc


#### Reference

参考相关开源项目：

https://github.com/tencentyun/cos-go-sdk-v5/

https://github.com/aliyun/aliyun-oss-go-sdk/oss

https://github.com/TencentCloud/kubernetes-csi-tencentcloud

https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner

https://github.com/yandex-cloud/k8s-csi-s3