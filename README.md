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
$ chmod +x aws-ps
```

3. Start your application with aws-ps
```
$(aws-ps -p /path/to/key -r ap-norhteast-1) && run.sh
```

## Example Dockerfile

- run.sh
    ```
    #!/bin/sh
    env
    
    echo "OK"
    ```
- Dockerfile
    ```
    FROM amazonlinux:2017.09
    
    RUN yum -y install wget tar
    
    ADD run.sh /run.sh
    
    RUN chmod +x /run.sh
    
    RUN wget https://github.com/hajimeni/aws-parameter-store-helper/releases/download/v0.1.0/aws-parameter-store-helper-linux-amd64.tar.gz \
     && tar xfz aws-parameter-store-helper-linux-amd64.tar.gz \
     && chmod +x aws-ps
     
    CMD $(aws-ps -p $AWS_PS_PATH -r $AWS_REGION) && run.sh
    ```
- build and run
    ```
    docker build -t app .
    docker run -e AWS_ACCESSK_KEY_ID=xxxx -e AWS_SECRET_ACCESS_KEY=yyyy -e AWS_PS_PATH=/path/to/key -e $AWS_REGION=ap-northeast-1 app
    ```
- then you can watch environment values such as below
    ```
    ...
    ENV_KEY_1=value1
    ENV_KEY_2=value2
    ...
    ```

## Commands

- load
    load aws ssm parameter stored values
- help
    help command

```
$ aws-ps help
AWS parameter store export helper:

ex)
  $ aws-ps load -p /path/to/key
  > export ENV_KEY=value;ENV_KEY_2=valu2

usage)
  $ eval $(AWS_REGION=ap-northeast-1 aws-ps load -p /path/to/key)
  then aws-pws will export environment parameters fetched from AWS Parameter Store:

Usage:
  aws-ps [command]

Available Commands:
  help        Help about any command
  load        load stored parameter then export formatted string

Flags:
  -h, --help   help for aws-ps

Use "aws-ps [command] --help" for more information about a command.
```
    
### `load` command Options

```
$ aws-ps load --help
Usage:
  aws-ps load [flags]

Flags:
  -d, --delimiter string   Delimiter each keys (default ";")
  -h, --help               help for load
  -p, --path string        Parameter Store Path, must starts with '/'
      --prefix string      Parameter Store Prefix. export KEY is removed prefix
  -r, --region string      AWS SDK region
  -t, --template string    export format template(Go Template) (default "export {{ .Name }}='{{ .Value }}'")
```

#### `-d` delimiter

ex)
```
$ aws-ps load -p /path/to/key -d ':'
export KEY_1=value1:export KEY_2=value2
```

#### `-t` template

ex)
```
$ aws-ps load -p /path/to/key -t '-D{{ .Name }}={{ .Value }}' -d ' '
-DKEY_1=value1 -DKEY_2=value2
```

#### `-p` path

starts with `/`

ex)
```
# paramete store
/path/to/key/KEY_1 -> value1
/path/to/key/KEY_2 -> value2

$ aws-ps load -p /path/to/key
export KEY_1=value1:export KEY_2=value2
```

#### `--prefix` prefix
 
`aws-ps` exports removed prefix keys.

ex)
```
# paramete store
path.to.key.KEY_1 -> value1
path.to.key.KEY_2 -> value2

$ aws-ps load --prefix path.to.key.
export KEY_1=value1:export KEY_2=value2
```

## How to build

1. Clone this repository
1. `go get -u github.com/golang/dep/cmd/dep`
1. `dep ensure`
1. run `./build.sh`
