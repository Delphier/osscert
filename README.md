# osscert
A simple Golang application to obtain/renew SSL certificates from Let's Encrypt and update to Aliyun OSS, supports running locally and deploy to Aliyun FC(Function Compute).

## Usage
Set environment variables:
| Name                   | Description                                    |
|------------------------|------------------------------------------------|
| OSS_ACCESS_KEY_ID      | Aliyun access key                              |
| OSS_ACCESS_KEY_SECRET  | Aliyun access secret                           |
| OSS_REGION             | Aliyun region ID: cn-hangzhou, cn-beijing, ... |
| OSS_BUCKET             | Bucket name                                    |
| ACME_EMAIL             | Email for ACME registration                    |

### Run locally:
```
osscert.exe
```

### Deploy to Aliyun Function Compute:
Upload ```osscert-fc.zip``` to FC.
