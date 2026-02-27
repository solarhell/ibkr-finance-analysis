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

#### 1. Personal Access Token

访问 [IBKR Portal](https://portal.ibkr.info/) 生成 Personal Access Token：
1. 登录后进入 **Settings → Account Settings**
2. 找到 **Access** 部分
3. 点击 **Create** 生成新的 Token
4. 复制生成的 Token 到配置文件

#### 2. Flex Query 配置

访问 [IBKR Flex Queries](https://www.interactivebrokers.com.hk/AccountManagement/AmAuthentication?action=FlexQueries) 创建查询：

**推荐配置（以 AllData365 为例）：**

| 配置项 | 值 |
|--------|-----|
| 查询名称 | AllData365 |
| 时间范围 | Last 365 Calendar Days |
| 输出格式 | **XML** (必须选择 XML，程序不支持 CSV 或其他格式) |

**⚠️ 重要：格式配置**

在 **Delivery Configuration** 部分，必须选择 **Format: XML**

- ✅ **XML** - 程序只能解析 XML 格式
- ❌ CSV - 不支持
- ❌ JSON - 不支持

**Date/Time 配置建议：**
- Date Format: **yyyy-MM-dd**（默认即可）
- Time Format: **HH:mm:ss TimeZone**（默认即可）
- Date/Time Separator: **; (semi-colon)**（默认即可）

**必需的 Sections（请全部勾选）：**

- **Cash Report** - 现金流水、佣金、股息汇总
- **Cash Transactions** - 股息、预扣税、利息等明细
- **Open Positions** - 当前持仓
- **Trades** - 交易记录
- **Transfers** - 转账记录

**Trades Section 配置建议：**
- Options: 选择 **Symbol Summary** 或 **Execution**
- 勾选 **Include Canceled Trades**: No

**Cash Transactions Section 配置建议：**
- Options: 勾选 **Dividends**, **Payment in Lieu of Dividends**, **Withholding Tax**, **Deposits & Withdrawals**

创建完成后，复制 **Query ID**（如 `1417381`）到配置文件。

#### 3. 账户 ID

在 IBKR Portal 的 **Account Summary** 或账户管理页面可以找到你的 Account ID（如 `U17389751`）。

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
