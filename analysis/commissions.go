package analysis

import (
	"fmt"
	"sort"

	"github.com/solarhell/finance-analysis/flex"
)

type SymbolCommission struct {
	Symbol     string
	Category   string
	Commission float64
	Trades     int
}

type CommissionReport struct {
	BySymbol       []SymbolCommission
	ByCategory     map[string]float64
	TotalComm      float64
	TotalTrades    int
}

func AnalyzeCommissions(statements []flex.FlexStatement, from, to string) *CommissionReport {
	symbolMap := make(map[string]*SymbolCommission)
	catMap := make(map[string]float64)

	var totalComm float64
	var totalTrades int

	for _, stmt := range statements {
		for _, t := range stmt.Trades {
			if !inDateRange(t.TradeDate, from, to) {
				continue
			}

			totalComm += t.Commission
			totalTrades++
			catMap[t.AssetCategory] += t.Commission

			sc, ok := symbolMap[t.Symbol]
			if !ok {
				sc = &SymbolCommission{Symbol: t.Symbol, Category: t.AssetCategory}
				symbolMap[t.Symbol] = sc
			}
			sc.Commission += t.Commission
			sc.Trades++
		}
	}

	report := &CommissionReport{
		ByCategory:  catMap,
		TotalComm:   totalComm,
		TotalTrades: totalTrades,
	}

	for _, sc := range symbolMap {
		report.BySymbol = append(report.BySymbol, *sc)
	}
	sort.Slice(report.BySymbol, func(i, j int) bool {
		return report.BySymbol[i].Commission < report.BySymbol[j].Commission // 佣金为负数，按绝对值排序
	})

	return report
}

func PrintCommissionReport(r *CommissionReport) {
	fmt.Println("═══ 佣金统计 ═══")
	fmt.Printf("总佣金:     %.2f\n", r.TotalComm)
	fmt.Printf("总交易数:   %d\n", r.TotalTrades)
	if r.TotalTrades > 0 {
		fmt.Printf("平均佣金:   %.2f\n", r.TotalComm/float64(r.TotalTrades))
	}
	fmt.Println()

	if len(r.ByCategory) > 0 {
		fmt.Println("── 按资产类别 ──")
		printTable(
			[]string{"类别", "佣金"},
			func() [][]string {
				var rows [][]string
				for cat, comm := range r.ByCategory {
					rows = append(rows, []string{cat, fmt.Sprintf("%.2f", comm)})
				}
				return rows
			}(),
		)
	}

	if len(r.BySymbol) > 0 {
		fmt.Println("── 按标的 ──")
		printTable(
			[]string{"标的", "类别", "佣金", "交易数"},
			func() [][]string {
				var rows [][]string
				for _, s := range r.BySymbol {
					rows = append(rows, []string{
						s.Symbol,
						s.Category,
						fmt.Sprintf("%.2f", s.Commission),
						fmt.Sprintf("%d", s.Trades),
					})
				}
				return rows
			}(),
		)
	}
}
