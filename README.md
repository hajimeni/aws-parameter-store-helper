aws-parameter-store-helper
----------------
[![Build Status](https://travis-ci.org/hajimeni/aws-parameter-store-helper.svg?branch=master)](https://travis-ci.org/hajimeni/aws-parameter-store-helper)

## Latest version

- `v0.4.0`
  - [![Build Status](https://travis-ci.org/hajimeni/aws-parameter-store-helper.svg?branch=v0.4.0)](https://travis-ci.org/hajimeni/aws-parameter-store-helper)
  - [Download(for mac OS X)](https://github.com/hajimeni/aws-parameter-store-helper/releases/download/v0.4.0/aws-ps-darwin-amd64.tar.gz)
  - [Download(for Linux)](https://github.com/hajimeni/aws-parameter-store-helper/releases/download/v0.4.0/aws-ps-linux-amd64.tar.gz)

## Usage

1. Add parameter to [Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html) using hierarchy in names:
```
$ aws ssm put-parameter --name /path/to/key/ENV_KEY_1 --value "value1" --type SecureString --key-id "alias/aws/ssm" --region ap-northeast-1
$ aws ssm put-parameter --name /path/to/key/ENV_KEY_2 --value "value2" --type SecureString --key-id "alias/aws/ssm" --region ap-northeast-1
```

2. Go to the [Releases Page](https://github.com/hajimeni/aws-parameter-store-helper/releases) and download the binary for your OS.
```
$ wget https://github.com/hajimeni/aws-parameter-store-helper/releases/download/v0.4.0/aws-parameter-store-helper-linux-amd64.tar.gz
$ tar xfz aws-parameter-store-helper-linux-amd64.tar.gz
$ chmod +x aws-ps
```

3. Start your application with aws-ps
```
$(aws-ps load -p /path/to/key -r ap-norhteast-1) && run.sh
```
or  
```
eval $(aws-ps load -p /path/to/key -r ap-northeast-1)
./run.sh
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

    RUN wget https://github.com/hajimeni/aws-parameter-store-helper/releases/download/v0.4.0/aws-parameter-store-helper-linux-amd64.tar.gz \
     && tar xfz aws-parameter-store-helper-linux-amd64.tar.gz \
     && chmod +x aws-ps

    CMD $(aws-ps load -p $AWS_PS_PATH -r $AWS_REGION) && run.sh
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
    OK
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
load stored parameter then export formatted string

Usage:
  aws-ps load [flags]

Flags:
  -d, --delimiter string           Delimiter each keys (default ";")
  -h, --help                       help for load
      --no-quote-shell             No quote shell characters
  -p, --path stringSlice           Parameter Store Path, must starts with '/'
      --prefix stringSlice         Parameter Store Prefix. export KEY is removed prefix
      --quote-shell                Quote shell characters (default true)
      --recursive                  Load recursive Parameter Store Path, '/' is escaped by escape-slash parameter
  -r, --region string              AWS SDK region
      --replace-key-value string   Replace parameter key each replace-keys characters to this value (default "_")
      --replace-keys string        Replace parameter key characters to replace-key-value (default "-/")
  -t, --template string            export format template(Go Template) (default "export {{ .Name }}=\"{{ .Value }}\"")
  -u, --uppercase-key              To upper case each parameter key
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

#### `-p` path (multiple)

starts with `/`

ex)
```
# paramete store
/path/to/key/KEY_1 -> value1
/path/to/key/KEY_2 -> value2
/path/to/hoge/KEY_3 -> value3
/path/to/hoge/KEY_4 -> value4
/path/to/hoge/KEY_1 -> value5

$ aws-ps load -p /path/to/key
export KEY_1=value1;export KEY_2=value2

# multiple path
$ aws-ps load -p /path/to/key -p /path/to/hoge
export KEY_1=value5;export KEY_2=value2;export KEY_3=value3;export KEY_4=value4

```

#### `--prefix` prefix (multiple)

`aws-ps` exports removed prefix keys.

ex)
```
# paramete store
path.to.key.KEY_1 -> value1
path.to.key.KEY_2 -> value2

$ aws-ps load --prefix path.to.key.
export KEY_1=value1;export KEY_2=value2
```

#### `-u` uppercase-key

To upper case each parameter keys

ex)
```
# paramete store
path.to.key.key_1 -> value1
path.to.key.KEY_2 -> value2

$ aws-ps load --prefix path.to.key. -u
export KEY_1=value1;export KEY_2=value2
```

#### `--recursive` recursive

load parameter store recursive by path
use with `--path` parameter

ex)
```
# paramete store
/path/to/key/KEY_1 -> value1
/path/to/key/recursive/KEY_2 -> value2

# recurisve
$ aws-ps load -p /path/to/key --recursive
export KEY_1=value1;export recursive_KEY_2=value2

# no-recursive(default)
$ aws-ps load -p /path/to/key
export KEY_1=value1

```

#### `--quote-shell` (default=true) `--no-quote-shell` (default=false)

quote variable shell specific characters.

ex)
```
# paramete store
/path/to/key/KEY_1 -> value1
/path/to/key/KEY_2 -> a[b+c="$\

## --quote-shell
$ aws-ps load -p /path/to/key
export KEY1="value1";export KEY2="a\[b+c=\"\$\\"

## --no-quote-shell
$ aws-ps load -p /path/to/key --no-quote-shell
export KEY1="value1";export KEY2="a[b+c="$\"
```

## How to build

1. Clone this repository
1. `go get -u github.com/golang/dep/cmd/dep`
1. `dep ensure`
1. run `./build.sh`
