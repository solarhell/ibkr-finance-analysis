package analysis

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// normalizeDate 将 YYYY-MM-DD 格式统一转换为 YYYYMMDD
func normalizeDate(date string) string {
	if len(date) == 10 && date[4] == '-' && date[7] == '-' {
		return date[:4] + date[5:7] + date[8:10]
	}
	return date
}

// inDateRange 检查日期是否在范围内（支持 YYYYMMDD 和 YYYY-MM-DD 格式）
func inDateRange(date, from, to string) bool {
	d := normalizeDate(date)
	if from != "" && d < normalizeDate(from) {
		return false
	}
	if to != "" && d > normalizeDate(to) {
		return false
	}
	return true
}

// printTable 用 tabwriter 打印表格
func printTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// 每个字段后加 \t
	for _, h := range headers {
		fmt.Fprintf(w, "%s\t", h)
	}
	fmt.Fprintln(w)

	for range headers {
		fmt.Fprintf(w, "──────\t")
	}
	fmt.Fprintln(w)

	for _, row := range rows {
		for _, cell := range row {
			fmt.Fprintf(w, "%s\t", cell)
		}
		fmt.Fprintln(w)
	}
	w.Flush()
	fmt.Println()
}
