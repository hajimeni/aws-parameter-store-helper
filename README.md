aws-parameter-store-helper
----------------

## Usage

1. Add parameter to [Parameter Store](https://console.aws.amazon.com/ec2/v2/home#Parameters:) using hierarchy in names:
```
$ aws ssm put-parameter --name /path/to/key/ENV_KEY_1 --value "value1" --type SecureString --key-id "alias/aws/ssm" --region ap-northeast-1
$ aws ssm put-parameter --name /path/to/key/ENV_KEY_2 --value "value2" --type SecureString --key-id "alias/aws/ssm" --region ap-northeast-1
```

2. Go to the [Releases Page](/hajimeni/aws-parameter-store-helper/releases) and download the binary for your OS.
```
$ wget https://github.com/hajimeni/aws-parameter-store-helper/releases/download/v0.1.0/aws-parameter-store-helper-linux-amd64.tar.gz
$ tar xfz aws-parameter-store-helper-linux-amd64.tar.gz
```

3. Start your application with aws-ps
```
$ eval(aws-ps -p /path/to/key -r ap-norhteast-1) && run.sh
```

## Example Dockerfile

```
...

ENV AWS_PS_PATH
RUN wget https://github.com/hajimeni/aws-parameter-store-helper/releases/download/v0.1.0/aws-parameter-store-helper-linux-amd64.tar.gz \
 && tar xfz aws-parameter-store-helper-linux-amd64.tar.gz \
 && chmod +x aws-ps
 
CMD eval $(aws-ps -p $AWS_PS_PATH) && run.sh
```

```
docker build -t app
docker run -e AWS_ACCESSK_KEY_ID=xxxx -e AWS_SECRET_ACCESS_KEY=yyyy -e AWS_PS_PATH=/path/to/key app
```