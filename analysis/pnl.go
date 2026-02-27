package analysis

import (
	"fmt"
	"math"
	"sort"

	"github.com/solarhell/ibkr-finance-analysis/flex"
)

type SymbolPnL struct {
	Symbol      string
	RealizedPnL float64
	Trades      int
	Wins        int
	Commission  float64
}

type PeriodPnL struct {
	Period      string
	RealizedPnL float64
	Trades      int
	Commission  float64
}

type PnLReport struct {
	BySymbol    []SymbolPnL
	ByMonth     []PeriodPnL
	TotalPnL    float64
	TotalTrades int
	WinRate     float64
	TotalComm   float64
}

// fifoLot 表示一个持仓批次
type fifoLot struct {
	qty      float64
	unitCost float64 // 每单位成本（多头买入成本 / 空头收到的权利金），恒为正
	isShort  bool
}

// computeFIFOPnL 用 FIFO 方法计算每个标的的已实现盈亏
// 处理多头（先买后卖）和空头（先卖后到期/平仓）两种情况
func computeFIFOPnL(trades []flex.Trade) map[string]float64 {
	sorted := make([]flex.Trade, len(trades))
	copy(sorted, trades)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].DateTime < sorted[j].DateTime
	})

	openLots := make(map[string][]fifoLot)
	pnl := make(map[string]float64)

	for _, t := range sorted {
		sym := t.Symbol
		qty := t.Quantity
		netCash := t.Proceeds + t.Commission // Commission 为负数

		if t.TransactionType == "BookTrade" {
			// 期权到期交割，以 0 平仓
			closeQty := math.Abs(qty)
			remaining := closeQty
			for remaining > 1e-9 && len(openLots[sym]) > 0 {
				l := &openLots[sym][0]
				match := math.Min(l.qty, remaining)
				if l.isShort {
					pnl[sym] += match * l.unitCost // 保住了权利金
				} else {
					pnl[sym] -= match * l.unitCost // 多头到期归零，损失买入成本
				}
				l.qty -= match
				remaining -= match
				if l.qty < 1e-9 {
					openLots[sym] = openLots[sym][1:]
				}
			}
		} else if qty > 0 {
			// 买入：先平空头，再开多头
			remaining := qty
			for remaining > 1e-9 && len(openLots[sym]) > 0 && openLots[sym][0].isShort {
				l := &openLots[sym][0]
				match := math.Min(l.qty, remaining)
				costToClosePerUnit := (-netCash) / qty
				pnl[sym] += match*l.unitCost - match*costToClosePerUnit
				l.qty -= match
				remaining -= match
				if l.qty < 1e-9 {
					openLots[sym] = openLots[sym][1:]
				}
			}
			if remaining > 1e-9 {
				unitCost := (-netCash) / qty
				openLots[sym] = append(openLots[sym], fifoLot{remaining, unitCost, false})
			}
		} else if qty < 0 {
			// 卖出：先平多头，再开空头
			sellQty := -qty
			remaining := sellQty
			for remaining > 1e-9 && len(openLots[sym]) > 0 && !openLots[sym][0].isShort {
				l := &openLots[sym][0]
				match := math.Min(l.qty, remaining)
				proceedsPerUnit := netCash / sellQty
				pnl[sym] += match*proceedsPerUnit - match*l.unitCost
				l.qty -= match
				remaining -= match
				if l.qty < 1e-9 {
					openLots[sym] = openLots[sym][1:]
				}
			}
			if remaining > 1e-9 {
				unitPremium := netCash / sellQty
				openLots[sym] = append(openLots[sym], fifoLot{remaining, unitPremium, true})
			}
		}
	}

	return pnl
}

// computeCostBasisFromTrades 返回每个标的当前持仓的 FIFO 总成本（仍持有的批次）
func computeCostBasisFromTrades(statements []flex.FlexStatement) map[string]float64 {
	var allTrades []flex.Trade
	for _, stmt := range statements {
		allTrades = append(allTrades, stmt.Trades...)
	}

	sorted := make([]flex.Trade, len(allTrades))
	copy(sorted, allTrades)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].DateTime < sorted[j].DateTime
	})

	openLots := make(map[string][]fifoLot)

	for _, t := range sorted {
		sym := t.Symbol
		qty := t.Quantity
		netCash := t.Proceeds + t.Commission

		if t.TransactionType == "BookTrade" {
			closeQty := math.Abs(qty)
			remaining := closeQty
			for remaining > 1e-9 && len(openLots[sym]) > 0 {
				l := &openLots[sym][0]
				match := math.Min(l.qty, remaining)
				l.qty -= match
				remaining -= match
				if l.qty < 1e-9 {
					openLots[sym] = openLots[sym][1:]
				}
			}
		} else if qty > 0 {
			remaining := qty
			for remaining > 1e-9 && len(openLots[sym]) > 0 && openLots[sym][0].isShort {
				l := &openLots[sym][0]
				match := math.Min(l.qty, remaining)
				l.qty -= match
				remaining -= match
				if l.qty < 1e-9 {
					openLots[sym] = openLots[sym][1:]
				}
			}
			if remaining > 1e-9 {
				unitCost := (-netCash) / qty
				openLots[sym] = append(openLots[sym], fifoLot{remaining, unitCost, false})
			}
		} else if qty < 0 {
			sellQty := -qty
			remaining := sellQty
			for remaining > 1e-9 && len(openLots[sym]) > 0 && !openLots[sym][0].isShort {
				l := &openLots[sym][0]
				match := math.Min(l.qty, remaining)
				l.qty -= match
				remaining -= match
				if l.qty < 1e-9 {
					openLots[sym] = openLots[sym][1:]
				}
			}
			if remaining > 1e-9 {
				unitPremium := netCash / sellQty
				openLots[sym] = append(openLots[sym], fifoLot{remaining, unitPremium, true})
			}
		}
	}

	result := make(map[string]float64)
	for sym, lots := range openLots {
		var totalCost float64
		for _, l := range lots {
			if !l.isShort {
				totalCost += l.qty * l.unitCost
			}
		}
		if totalCost > 0 {
			result[sym] = totalCost
		}
	}
	return result
}

func AnalyzePnL(statements []flex.FlexStatement, from, to string) *PnLReport {
	// 收集符合日期范围的所有交易
	var filteredTrades []flex.Trade
	for _, stmt := range statements {
		for _, t := range stmt.Trades {
			if !inDateRange(normalizeDate(t.TradeDate), from, to) {
				continue
			}
			filteredTrades = append(filteredTrades, t)
		}
	}

	// FIFO 计算已实现盈亏
	fifoResult := computeFIFOPnL(filteredTrades)

	// 按标的统计佣金和交易次数（以平仓 ExchTrade 为准）
	symbolMap := make(map[string]*SymbolPnL)
	monthMap := make(map[string]*PeriodPnL)

	var totalComm float64
	for _, t := range filteredTrades {
		totalComm += t.Commission

		sp, ok := symbolMap[t.Symbol]
		if !ok {
			sp = &SymbolPnL{Symbol: t.Symbol}
			symbolMap[t.Symbol] = sp
		}
		sp.Commission += t.Commission

		// 只统计平仓 ExchTrade 的交易次数
		if t.TransactionType != "BookTrade" && (t.Quantity < 0 || t.OpenCloseInd == "C" || t.OpenCloseInd == "C;") {
			nd := normalizeDate(t.TradeDate)
			month := ""
			if len(nd) >= 6 {
				month = nd[:6]
			}
			mp, ok := monthMap[month]
			if !ok {
				mp = &PeriodPnL{Period: month}
				monthMap[month] = mp
			}
			mp.Trades++
		}
	}

	// 填入 FIFO 盈亏
	var totalPnL float64
	var totalTrades, totalWins int

	for sym, symPnL := range fifoResult {
		if math.Abs(symPnL) < 1e-6 {
			continue // 跳过仍持有的标的（盈亏为 0）
		}
		sp, ok := symbolMap[sym]
		if !ok {
			sp = &SymbolPnL{Symbol: sym}
			symbolMap[sym] = sp
		}
		sp.RealizedPnL = symPnL
		sp.Trades = 1
		if symPnL > 0 {
			sp.Wins = 1
			totalWins++
		}
		totalTrades++
		totalPnL += symPnL

		// 月度汇总用第一个平仓交易的月份
		for _, t := range filteredTrades {
			if t.Symbol != sym {
				continue
			}
			isSell := t.Quantity < 0
			isClose := t.TransactionType == "BookTrade" || t.OpenCloseInd == "C" || t.OpenCloseInd == "C;"
			if isSell || isClose {
				nd := normalizeDate(t.TradeDate)
				month := ""
				if len(nd) >= 6 {
					month = nd[:6]
				}
				mp, ok := monthMap[month]
				if !ok {
					mp = &PeriodPnL{Period: month}
					monthMap[month] = mp
				}
				mp.RealizedPnL += symPnL
				break
			}
		}
	}

	report := &PnLReport{
		TotalPnL:    totalPnL,
		TotalTrades: totalTrades,
		TotalComm:   totalComm,
	}
	if totalTrades > 0 {
		report.WinRate = float64(totalWins) / float64(totalTrades) * 100
	}

	for _, sp := range symbolMap {
		report.BySymbol = append(report.BySymbol, *sp)
	}
	sort.Slice(report.BySymbol, func(i, j int) bool {
		return report.BySymbol[i].RealizedPnL > report.BySymbol[j].RealizedPnL
	})

	for _, mp := range monthMap {
		report.ByMonth = append(report.ByMonth, *mp)
	}
	sort.Slice(report.ByMonth, func(i, j int) bool {
		return report.ByMonth[i].Period < report.ByMonth[j].Period
	})

	return report
}

func formatMonth(yyyymm string) string {
	if len(yyyymm) >= 6 {
		return yyyymm[:4] + "-" + yyyymm[4:6]
	}
	return yyyymm
}

func PrintPnLReport(r *PnLReport) {
	fmt.Println("═══ 盈亏分析 ═══")
	fmt.Printf("已实现盈亏: %.2f\n", r.TotalPnL)
	fmt.Printf("总交易数:   %d\n", r.TotalTrades)
	fmt.Printf("胜率:       %.1f%%\n", r.WinRate)
	fmt.Printf("总佣金:     %.2f\n", r.TotalComm)
	fmt.Println()

	if len(r.BySymbol) > 0 {
		fmt.Println("── 按标的 ──")
		printTable(
			[]string{"标的", "已实现P&L", "交易数", "胜率", "佣金"},
			func() [][]string {
				var rows [][]string
				for _, s := range r.BySymbol {
					wr := 0.0
					if s.Trades > 0 {
						wr = float64(s.Wins) / float64(s.Trades) * 100
					}
					rows = append(rows, []string{
						s.Symbol,
						fmt.Sprintf("%.2f", s.RealizedPnL),
						fmt.Sprintf("%d", s.Trades),
						fmt.Sprintf("%.0f%%", wr),
						fmt.Sprintf("%.2f", s.Commission),
					})
				}
				return rows
			}(),
		)
	}

	if len(r.ByMonth) > 0 {
		fmt.Println("── 按月份 ──")
		printTable(
			[]string{"月份", "已实现P&L", "交易数", "佣金"},
			func() [][]string {
				var rows [][]string
				for _, m := range r.ByMonth {
					rows = append(rows, []string{
						formatMonth(m.Period),
						fmt.Sprintf("%.2f", m.RealizedPnL),
						fmt.Sprintf("%d", m.Trades),
						fmt.Sprintf("%.2f", m.Commission),
					})
				}
				return rows
			}(),
		)
	}
}
