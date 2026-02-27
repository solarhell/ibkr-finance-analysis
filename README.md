# IBKR Finance Analysis

Interactive Brokers (IBKR) 交易数据分析和报告生成工具。

## 功能

- 通过 IBKR Flex API 获取交易数据
- 分析交易记录、佣金、股息收入、盈亏等
- 生成 Markdown 格式的分析报告

## 配置

1. 复制配置示例文件：
```bash
cp config.example.toml config.toml
```

2. 编辑 `config.toml`，填入你的 IBKR API 信息：
```toml
token = "your_personal_access_token_here"
account_id = "your_account_id_here"

[queries]
trades = "your_query_id_here"
```

### 获取 API 凭证

- **Token**: 访问 [IBKR Portal](https://portal.ibkr.info/) 生成 Personal Access Token
- **Query ID**: 在 IBKR Portal 的 Flex 查询中创建查询并获取 ID

## 使用

```bash
# 运行分析
go run main.go
```

分析报告将保存在 `data/` 目录下。

## 项目结构

```
.
├── main.go           # 程序入口
├── config.go         # 配置加载
├── flex/             # IBKR Flex API 客户端
├── analysis/         # 数据分析模块
└── data/             # 数据存储目录
```
