# gomodule

## 功能

- 模块化加载
- 配置文件自动加载
- 自动给模块的配置结构体赋值
- 加载本地配置文件、网络配置文件，可以将etcd、consul的键值作为配置项
- 对配置文件动态加载，实时修改实时生效，无需重启进程
- 支持logrus日志库
- 模块生命周期统一管理
- 支持服务注册、卸载、开始、停止、重启
- 支持cobra.Command库，方便实现命令行开发

详细使用流程可参考`examples`下的demo

## Demo 编译运行

### 标准模式启动

```bash
cd examples
go run simplemodule.go
```

```log
INFO[0000] get module named: ConfigCenter                module=manager
INFO[0000] register default module: ConfigCenter         module=manager
INFO[0000] get module named: SimpleModule                module=manager
INFO[0000] register module: SimpleModule                 module=manager
INFO[0000] register default module: syservice            module=manager
INFO[0000] register default module: config               module=manager
INFO[0000] register default module: logger               module=manager
INFO[0000] launch manager, modules: 5                    module=manager
INFO[0000] launch manager, default modules: 4            module=manager
INFO[0000] init configcenter module                     
INFO[0000] init syservice module                         module=syserver
INFO[0000] init logger module                            module=logger
INFO[0000] init simple module                            module=simple
INFO[0000] init command                                 
INFO[0000] current user &{Uid:501 Gid:20 Username:root Name:root HomeDir:/Users/root}  module=syserver
INFO[0000] init config module                            module=config
INFO[0000] init command                                  module=simple
INFO[0000] root command                                 

  SERVER  | DOMAIN  | ADDRESS | METHOD | ROUTE |                                        HANDLER                                        | MIDDLEWARE  
----------|---------|---------|--------|-------|---------------------------------------------------------------------------------------|-------------
  default | default | :9990   | ALL    | /     | github.com/let-light/gomodule/examples/configcenter.(*ConfigCenter).RootCommand.func1 |             
----------|---------|---------|--------|-------|---------------------------------------------------------------------------------------|-------------
 
2023-04-14 15:53:21.136 39560: http server started listening on [:9990] 
INFO[0001] config flags: {LocalFile: Consul: Etcd: RemoteFile: RemoteFileInterval:30}  module=config
INFO[0001] default config file: config.yml (./config.yml or ./config/config.yml)  module=config
INFO[0001] reload settings, modules:5                    module=config
INFO[0001] reload settings:config                        module=config
INFO[0001] reload settings:logger                        module=config
INFO[0001] reload settings:SimpleModule                  module=config
INFO[2023-04-14 15:53:22] root command                                  module=simple
INFO[2023-04-14 15:53:22] settings: hello world                         module=simple
INFO[2023-04-14 15:53:23] tick...                                       module=simple
INFO[2023-04-14 15:53:24] tick...                                       module=simple
```

### 命令行操作实例

```bash
cd examples
go run simplemodule.go desc
```

```log
INFO[0000] get module named: ConfigCenter                module=manager
INFO[0000] register default module: ConfigCenter         module=manager
INFO[0000] get module named: SimpleModule                module=manager
INFO[0000] register module: SimpleModule                 module=manager
INFO[0000] register default module: syservice            module=manager
INFO[0000] register default module: config               module=manager
INFO[0000] register default module: logger               module=manager
INFO[0000] launch manager, modules: 5                    module=manager
INFO[0000] launch manager, default modules: 4            module=manager
INFO[0000] init configcenter module                     
INFO[0000] init syservice module                         module=syserver
INFO[0000] init logger module                            module=logger
INFO[0000] init simple module                            module=simple
INFO[0000] init command                                 
INFO[0000] current user &{Uid:501 Gid:20 Username:root Name:root HomeDir:/Users/root}  module=syserver
INFO[0000] init config module                            module=config
INFO[0000] init command                                  module=simple
INFO[0000] simple module command run ...                 module=simple
```

### 服务注册

```bash
go build simplemodule
sudo ./simplemodule serv --ctrl=install --name=gomodule --workdir=/var/gomodule
```

```log
INFO[0000] get module named: ConfigCenter                module=manager
INFO[0000] register default module: ConfigCenter         module=manager
INFO[0000] get module named: SimpleModule                module=manager
INFO[0000] register module: SimpleModule                 module=manager
INFO[0000] register default module: syservice            module=manager
INFO[0000] register default module: config               module=manager
INFO[0000] register default module: logger               module=manager
INFO[0000] launch manager, modules: 5                    module=manager
INFO[0000] launch manager, default modules: 4            module=manager
INFO[0000] init configcenter module                     
INFO[0000] init syservice module                         module=syserver
INFO[0000] init logger module                            module=logger
INFO[0000] init simple module                            module=simple
INFO[0000] init command                                 
INFO[0000] current user &{Uid:0 Gid:0 Username:root Name:System Administrator HomeDir:/var/root}  module=syserver
INFO[0000] init config module                            module=config
INFO[0000] init command                                  module=simple
INFO[0000] service install success                       module=syserver
```

```bash
sudo ./simplemodule serv --ctrl=start --name=gomodule
```

```bash
sudo ./simplemodule serv --ctrl=restart --name=gomodule
```

```bash
sudo ./simplemodule serv --ctrl=stop --name=gomodule
```

```bash
sudo ./simplemodule serv --ctrl=uninstall --name=gomodule
```
