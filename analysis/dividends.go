package analysis

import (
	"fmt"
	"sort"

	"github.com/solarhell/finance-analysis/flex"
)

type SymbolDividend struct {
	Symbol       string
	Gross        float64
	Withholding  float64
	Net          float64
	Transactions int
}

type DividendReport struct {
	BySymbol       []SymbolDividend
	TotalGross     float64
	TotalWithhold  float64
	TotalNet       float64
	TotalCount     int
}

func AnalyzeDividends(statements []flex.FlexStatement, from, to string) *DividendReport {
	symbolMap := make(map[string]*SymbolDividend)

	var totalGross, totalWithhold float64
	var totalCount int

	for _, stmt := range statements {
		for _, ct := range stmt.CashTransactions {
			if !inDateRange(normalizeDate(ct.TradeDate), from, to) {
				continue
			}

			switch ct.Type {
			case "Dividends", "Payment In Lieu Of Dividends":
				totalGross += ct.Amount
				totalCount++
				sd, ok := symbolMap[ct.Symbol]
				if !ok {
					sd = &SymbolDividend{Symbol: ct.Symbol}
					symbolMap[ct.Symbol] = sd
				}
				sd.Gross += ct.Amount
				sd.Transactions++

			case "Withholding Tax":
				totalWithhold += ct.Amount // 通常为负数
				sd, ok := symbolMap[ct.Symbol]
				if !ok {
					sd = &SymbolDividend{Symbol: ct.Symbol}
					symbolMap[ct.Symbol] = sd
				}
				sd.Withholding += ct.Amount
			}
		}
	}

	report := &DividendReport{
		TotalGross:    totalGross,
		TotalWithhold: totalWithhold,
		TotalNet:      totalGross + totalWithhold, // withholding 为负数
		TotalCount:    totalCount,
	}

	for _, sd := range symbolMap {
		sd.Net = sd.Gross + sd.Withholding
		report.BySymbol = append(report.BySymbol, *sd)
	}
	sort.Slice(report.BySymbol, func(i, j int) bool {
		return report.BySymbol[i].Net > report.BySymbol[j].Net
	})

	return report
}

func PrintDividendReport(r *DividendReport) {
	fmt.Println("═══ 股息统计 ═══")
	fmt.Printf("总股息收入:   %.2f\n", r.TotalGross)
	fmt.Printf("预扣税:       %.2f\n", r.TotalWithhold)
	fmt.Printf("净股息收入:   %.2f\n", r.TotalNet)
	fmt.Printf("派息次数:     %d\n", r.TotalCount)
	fmt.Println()

	if len(r.BySymbol) > 0 {
		fmt.Println("── 按标的 ──")
		printTable(
			[]string{"标的", "总股息", "预扣税", "净收入", "次数"},
			func() [][]string {
				var rows [][]string
				for _, s := range r.BySymbol {
					rows = append(rows, []string{
						s.Symbol,
						fmt.Sprintf("%.2f", s.Gross),
						fmt.Sprintf("%.2f", s.Withholding),
						fmt.Sprintf("%.2f", s.Net),
						fmt.Sprintf("%d", s.Transactions),
					})
				}
				return rows
			}(),
		)
	}
}
