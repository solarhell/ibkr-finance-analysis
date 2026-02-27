package analysis

import (
	"fmt"
	"strings"
	"time"

	"github.com/solarhell/finance-analysis/flex"
)

func GenerateMarkdownReport(statements []flex.FlexStatement, from, to string) string {
	summary := AnalyzeSummary(statements, from, to)
	pnl := AnalyzePnL(statements, from, to)
	divs := AnalyzeDividends(statements, from, to)

	// 获取报告期间
	var periodFrom, periodTo string
	for _, stmt := range statements {
		if periodFrom == "" || stmt.FromDate < periodFrom {
			periodFrom = stmt.FromDate
		}
		if periodTo == "" || stmt.ToDate > periodTo {
			periodTo = stmt.ToDate
		}
	}

	var b strings.Builder

	// 标题
	b.WriteString("# IBKR 账户报告\n\n")
	b.WriteString(fmt.Sprintf("**报告期间：** %s — %s\n\n", formatDate(periodFrom), formatDate(periodTo)))
	b.WriteString(fmt.Sprintf("**生成时间：** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	b.WriteString("---\n\n")

	// 账户总值
	b.WriteString("## 账户总值\n\n")
	b.WriteString(fmt.Sprintf("| 项目 | 金额 (USD) |\n|------|----------:|\n"))
	b.WriteString(fmt.Sprintf("| 持仓市值 | %s |\n", fmtMoney(summary.TotalValue)))
	b.WriteString(fmt.Sprintf("| 现金余额 | %s |\n", fmtMoney(summary.CashBalance)))
	b.WriteString(fmt.Sprintf("| **账户总值** | **%s** |\n", fmtMoney(summary.AccountValue)))
	b.WriteString("\n")

	// 资金流动
	b.WriteString("## 资金流动\n\n")
	b.WriteString("| 项目 | 金额 (USD) |\n|------|----------:|\n")
	b.WriteString(fmt.Sprintf("| 总入金 | %s |\n", fmtMoney(summary.TotalDeposits)))
	b.WriteString(fmt.Sprintf("| 总出金 | %s |\n", fmtMoney(summary.TotalWithdrawals)))
	netFlow := summary.TotalDeposits + summary.TotalWithdrawals
	b.WriteString(fmt.Sprintf("| **净入金** | **%s** |\n", fmtMoney(netFlow)))
	b.WriteString("\n")

	// 收益总览
	totalReturn := summary.TotalRealPnL + summary.TotalUnrealPnL + summary.TotalDivNet + summary.TotalCommission
	returnPct := 0.0
	if summary.TotalDeposits > 0 {
		returnPct = totalReturn / summary.TotalDeposits * 100
	}
	b.WriteString("## 收益总览\n\n")
	b.WriteString("| 项目 | 金额 (USD) |\n|------|----------:|\n")
	b.WriteString(fmt.Sprintf("| 已实现盈亏 | %s |\n", fmtPnL(summary.TotalRealPnL)))
	b.WriteString(fmt.Sprintf("| 未实现盈亏 | %s |\n", fmtPnL(summary.TotalUnrealPnL)))
	b.WriteString(fmt.Sprintf("| 净股息收入 | %s |\n", fmtPnL(summary.TotalDivNet)))
	b.WriteString(fmt.Sprintf("| 佣金支出 | %s |\n", fmtPnL(summary.TotalCommission)))
	b.WriteString(fmt.Sprintf("| **综合收益** | **%s** |\n", fmtPnL(totalReturn)))
	if summary.TotalDeposits > 0 {
		b.WriteString(fmt.Sprintf("| 收益率（基于入金） | **%.2f%%** |\n", returnPct))
	}
	b.WriteString("\n")

	// 当前持仓
	if len(summary.Positions) > 0 {
		b.WriteString("## 当前持仓\n\n")
		b.WriteString("| 标的 | 数量 | 现价 | 成本价 | 市值 | 未实现P&L | 占比 |\n")
		b.WriteString("|------|-----:|-----:|-------:|-----:|----------:|-----:|\n")
		for _, p := range summary.Positions {
			pct := 0.0
			if summary.TotalValue > 0 {
				pct = p.PositionValue / summary.TotalValue * 100
			}
			b.WriteString(fmt.Sprintf("| %s | %.4g | %.2f | %.2f | %.2f | %s | %.1f%% |\n",
				p.Symbol,
				p.Position,
				p.MarkPrice,
				p.CostBasis,
				p.PositionValue,
				fmtPnL(p.UnrealizedPnL),
				pct,
			))
		}
		b.WriteString("\n")
	}

	// 已实现盈亏明细
	if len(pnl.BySymbol) > 0 {
		b.WriteString("## 已实现盈亏明细\n\n")
		b.WriteString(fmt.Sprintf("- 平仓交易数：%d 笔　胜率：%.1f%%　总佣金：%.2f\n\n",
			pnl.TotalTrades, pnl.WinRate, pnl.TotalComm))
		b.WriteString("| 标的 | 已实现P&L | 交易数 | 胜率 | 佣金 |\n")
		b.WriteString("|------|----------:|------:|-----:|-----:|\n")
		for _, s := range pnl.BySymbol {
			wr := 0.0
			if s.Trades > 0 {
				wr = float64(s.Wins) / float64(s.Trades) * 100
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %d | %.0f%% | %.2f |\n",
				s.Symbol, fmtPnL(s.RealizedPnL), s.Trades, wr, s.Commission))
		}
		b.WriteString("\n")
	}

	// 月度收益
	if len(pnl.ByMonth) > 0 {
		b.WriteString("## 月度收益\n\n")
		b.WriteString("| 月份 | 已实现P&L | 交易数 | 佣金 |\n")
		b.WriteString("|------|----------:|------:|-----:|\n")
		for _, m := range pnl.ByMonth {
			b.WriteString(fmt.Sprintf("| %s | %s | %d | %.2f |\n",
				formatMonth(m.Period), fmtPnL(m.RealizedPnL), m.Trades, m.Commission))
		}
		b.WriteString("\n")
	}

	// 股息明细（从 CashTransactions，若无则从 CashReport 汇总）
	b.WriteString("## 股息收入\n\n")
	if len(divs.BySymbol) > 0 {
		b.WriteString(fmt.Sprintf("- 总股息：%.2f　预扣税：%.2f　**净收入：%.2f**　派息次数：%d\n\n",
			divs.TotalGross, divs.TotalWithhold, divs.TotalNet, divs.TotalCount))
		b.WriteString("| 标的 | 税前股息 | 预扣税 | 净收入 | 次数 |\n")
		b.WriteString("|------|--------:|------:|------:|-----:|\n")
		for _, s := range divs.BySymbol {
			b.WriteString(fmt.Sprintf("| %s | %.2f | %.2f | %.2f | %d |\n",
				s.Symbol, s.Gross, s.Withholding, s.Net, s.Transactions))
		}
	} else {
		// 从 CashReport 汇总
		b.WriteString(fmt.Sprintf("净股息收入（含预扣税）：**%.2f USD**\n", summary.TotalDivNet))
		b.WriteString("\n> 详细股息明细需在 AllData365 Flex Query 中添加 Cash Transactions 段\n")
	}
	b.WriteString("\n")

	return b.String()
}

func fmtMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func fmtPnL(v float64) string {
	if v > 0 {
		return fmt.Sprintf("+%.2f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

func formatDate(d string) string {
	if len(d) == 8 {
		return d[:4] + "-" + d[4:6] + "-" + d[6:]
	}
	return d
}
