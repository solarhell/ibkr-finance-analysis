package analysis

import (
	"fmt"
	"sort"

	"github.com/solarhell/finance-analysis/flex"
)

type PositionSummary struct {
	Symbol         string
	Category       string
	Currency       string
	Position       float64
	MarkPrice      float64
	CostBasis      float64
	PositionValue  float64
	UnrealizedPnL  float64
}

type SummaryReport struct {
	Positions       []PositionSummary
	TotalValue      float64
	TotalUnrealPnL  float64
	TotalRealPnL    float64
	TotalDivNet     float64
	TotalCommission float64
	CashBalance     float64
	TotalDeposits   float64
	TotalWithdrawals float64
	AccountValue    float64 // 持仓 + 现金
}

func AnalyzeSummary(statements []flex.FlexStatement, from, to string) *SummaryReport {
	report := &SummaryReport{}

	// 用 FIFO 推算持仓成本价（当 OpenPosition 里的 costBasisPrice 为 0 时）
	costBySymbol := computeCostBasisFromTrades(statements)

	// 持仓概览
	for _, stmt := range statements {
		for _, op := range stmt.OpenPositions {
			costBasis := op.CostBasis
			if costBasis == 0 {
				if totalCost, ok := costBySymbol[op.Symbol]; ok && op.Position > 0 {
					costBasis = totalCost / op.Position
				}
			}
			unrealPnL := op.FifoPnlUnrealized
			if unrealPnL == 0 && costBasis > 0 {
				unrealPnL = op.PositionValue - costBasis*op.Position
			}
			ps := PositionSummary{
				Symbol:        op.Symbol,
				Category:      op.AssetCategory,
				Currency:      op.Currency,
				Position:      op.Position,
				MarkPrice:     op.MarkPrice,
				CostBasis:     costBasis,
				PositionValue: op.PositionValue,
				UnrealizedPnL: unrealPnL,
			}
			report.Positions = append(report.Positions, ps)
			report.TotalValue += op.PositionValue
			report.TotalUnrealPnL += unrealPnL
		}
	}

	sort.Slice(report.Positions, func(i, j int) bool {
		return report.Positions[i].PositionValue > report.Positions[j].PositionValue
	})

	// 从 CashReport 获取期间数据（BASE_SUMMARY 行）
	for _, stmt := range statements {
		for _, cr := range stmt.CashReport {
			if cr.Currency == "BASE_SUMMARY" {
				report.TotalCommission = cr.Commissions
				report.TotalDivNet = cr.Dividends + cr.WithholdingTax
				report.CashBalance = cr.EndingCash
				report.TotalDeposits = cr.Deposits
				report.TotalWithdrawals = cr.Withdrawals
			}
		}
	}

	// 已实现盈亏（从 Trades）
	pnl := AnalyzePnL(statements, from, to)
	report.TotalRealPnL = pnl.TotalPnL

	report.AccountValue = report.TotalValue + report.CashBalance

	return report
}

func PrintSummaryReport(r *SummaryReport) {
	fmt.Println("═══ 账户综合汇总 ═══")
	fmt.Println()
	fmt.Printf("账户总值:       %.2f\n", r.AccountValue)
	fmt.Printf("  持仓市值:     %.2f\n", r.TotalValue)
	fmt.Printf("  现金余额:     %.2f\n", r.CashBalance)
	fmt.Println()
	fmt.Printf("未实现盈亏:     %.2f\n", r.TotalUnrealPnL)
	fmt.Printf("已实现盈亏:     %.2f\n", r.TotalRealPnL)
	fmt.Printf("净股息收入:     %.2f\n", r.TotalDivNet)
	fmt.Printf("总佣金:         %.2f\n", r.TotalCommission)
	fmt.Printf("总入金:         %.2f\n", r.TotalDeposits)
	fmt.Printf("总出金:         %.2f\n", r.TotalWithdrawals)
	fmt.Println()

	if len(r.Positions) > 0 {
		fmt.Println("── 当前持仓 ──")
		printTable(
			[]string{"标的", "数量", "现价", "成本价", "市值", "未实现P&L"},
			func() [][]string {
				var rows [][]string
				for _, p := range r.Positions {
					rows = append(rows, []string{
						p.Symbol,
						fmt.Sprintf("%.0f", p.Position),
						fmt.Sprintf("%.2f", p.MarkPrice),
						fmt.Sprintf("%.2f", p.CostBasis),
						fmt.Sprintf("%.2f", p.PositionValue),
						fmt.Sprintf("%.2f", p.UnrealizedPnL),
					})
				}
				return rows
			}(),
		)
	}
}
