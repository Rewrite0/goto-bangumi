本项目是一个解析 RSS, 得到对应的种子文件, 并自动推送下载,通知,改名的工具
本项目的模块分为
1. RSS 解析模块: 负责解析 RSS 源, 获取最新的种子信息
- 将 RSS 源解析为种子信息, 并将其存储到数据库中
2. 种子下载模块: 抽象对应的下载器
- 目前实现 qbittorrent 下载器, cd2 还有待实现
- 下载器要提供下载, 移动和重命名功能
3. 通知模块 internal/notification: 负责将下载结果通知到指定的渠道, 如邮件,telegram
4. 配置模块 internal/conf : 负责读取和管理配置文件,
- 对配置检验
- 提供热重载
5. 日志模块 internal/logger : 负责记录程序运行日志
- 可以热重载日志等级
- 时间片轮转
6. 数据库模块 internal/download : 负责管理数据库连接和操作
7. 网络模块 internal/network : 负责网络请求, 提供基础的网络请求功能
8. 重命名模块 internal/rename : 按照配置中的重命名规则, 对下载完成的文件进行重命名
- 接收Bangumi 和 Torrent, 把种子生成对应的动漫名字
9. rss 模块 internal/rss : 负责解析 RSS 源, 获取最新的种子信息
10. scheduler 模块 internal/scheduler : 负责定时任务调度, 如定时检查 RSS 源, 定时清理数据库等
11. task 模块 internal/taskrunner : 负责管理任务队列
实现一个种子的生命周期轮转，从 rss 解析成为一个 torrent 后，进入到这里完成下载，重命名，通知等一系列操作
保证一个种子 url 为标准的去重，同一时间，一个种子只能进入一次
