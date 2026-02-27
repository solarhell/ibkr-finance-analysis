package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/solarhell/ibkr-finance-analysis/analysis"
	"github.com/solarhell/ibkr-finance-analysis/flex"
	"github.com/spf13/cobra"
)

var (
	flagFrom   string
	flagTo     string
	flagFormat string
	flagQuery  string
)

func main() {
	root := &cobra.Command{
		Use:   "ibkr",
		Short: "IBKR 交易记录分析工具",
	}

	root.PersistentFlags().StringVar(&flagFrom, "from", "", "起始日期 (YYYYMMDD)")
	root.PersistentFlags().StringVar(&flagTo, "to", "", "结束日期 (YYYYMMDD)")
	root.PersistentFlags().StringVar(&flagFormat, "format", "table", "输出格式: table, json")

	root.AddCommand(fetchCmd())
	root.AddCommand(analyzeCmd())
	root.AddCommand(syncCmd())
	root.AddCommand(reportCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func fetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "拉取 Flex Query 数据并保存为本地 XML",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig()
			if err != nil {
				return err
			}

			client := flex.NewClient(cfg.Token)

			queries := cfg.Queries
			if flagQuery != "" {
				// 只拉取指定的 query
				qid, ok := cfg.Queries[flagQuery]
				if !ok {
					return fmt.Errorf("未找到 query: %s (可用: %s)", flagQuery, availableQueries(cfg))
				}
				queries = map[string]string{flagQuery: qid}
			}

			for name, qid := range queries {
				resp, rawXML, err := client.FetchQuery(qid)
				if err != nil {
					return fmt.Errorf("拉取 %s 失败: %w", name, err)
				}

				filename := fmt.Sprintf("%s_%s.xml", name, time.Now().Format("20060102_150405"))
				path := filepath.Join(cfg.DataDir, filename)
				if err := os.WriteFile(path, rawXML, 0644); err != nil {
					return fmt.Errorf("保存文件失败: %w", err)
				}

				stmtCount := len(resp.FlexStatements)
				tradeCount := 0
				for _, s := range resp.FlexStatements {
					tradeCount += len(s.Trades)
				}
				fmt.Printf("✓ %s: 已保存到 %s (%d 账户, %d 笔交易)\n", name, filename, stmtCount, tradeCount)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&flagQuery, "query", "q", "", "指定拉取的 query 名称")
	return cmd
}

func analyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze [trades|dividends|commissions|summary]",
		Short: "分析已拉取的数据",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig()
			if err != nil {
				return err
			}

			statements, err := loadLatestData(cfg)
			if err != nil {
				return err
			}

			return runAnalysis(args[0], statements, flagFrom, flagTo, flagFormat)
		},
	}
	return cmd
}

func syncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [trades|dividends|commissions|summary]",
		Short: "拉取数据并分析（fetch + analyze）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig()
			if err != nil {
				return err
			}

			client := flex.NewClient(cfg.Token)
			var allStatements []flex.FlexStatement

			for name, qid := range cfg.Queries {
				resp, rawXML, err := client.FetchQuery(qid)
				if err != nil {
					return fmt.Errorf("拉取 %s 失败: %w", name, err)
				}

				filename := fmt.Sprintf("%s_%s.xml", name, time.Now().Format("20060102_150405"))
				path := filepath.Join(cfg.DataDir, filename)
				if err := os.WriteFile(path, rawXML, 0644); err != nil {
					return fmt.Errorf("保存文件失败: %w", err)
				}
				fmt.Printf("✓ %s: 已保存到 %s\n", name, filename)

				allStatements = append(allStatements, resp.FlexStatements...)
			}

			fmt.Println()
			return runAnalysis(args[0], allStatements, flagFrom, flagTo, flagFormat)
		},
	}
}

func runAnalysis(mode string, statements []flex.FlexStatement, from, to, format string) error {
	switch mode {
	case "trades", "pnl":
		r := analysis.AnalyzePnL(statements, from, to)
		if format == "json" {
			return printJSON(r)
		}
		analysis.PrintPnLReport(r)

	case "dividends":
		r := analysis.AnalyzeDividends(statements, from, to)
		if format == "json" {
			return printJSON(r)
		}
		analysis.PrintDividendReport(r)

	case "commissions":
		r := analysis.AnalyzeCommissions(statements, from, to)
		if format == "json" {
			return printJSON(r)
		}
		analysis.PrintCommissionReport(r)

	case "summary":
		r := analysis.AnalyzeSummary(statements, from, to)
		if format == "json" {
			return printJSON(r)
		}
		analysis.PrintSummaryReport(r)

	default:
		return fmt.Errorf("未知分析类型: %s (可用: trades, dividends, commissions, summary)", mode)
	}
	return nil
}

func loadLatestData(cfg *Config) ([]flex.FlexStatement, error) {
	var allStatements []flex.FlexStatement

	for name := range cfg.Queries {
		pattern := filepath.Join(cfg.DataDir, name+"_*.xml")
		matches, err := filepath.Glob(pattern)
		if err != nil || len(matches) == 0 {
			fmt.Printf("跳过 %s: 无本地数据，请先执行 fetch\n", name)
			continue
		}
		// 文件名含时间戳，字典序最大即最新
		latest := matches[0]
		for _, m := range matches[1:] {
			if filepath.Base(m) > filepath.Base(latest) {
				latest = m
			}
		}

		data, err := os.ReadFile(latest)
		if err != nil {
			return nil, fmt.Errorf("读取 %s 失败: %w", latest, err)
		}

		var resp flex.FlexQueryResponse
		if err := xml.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("解析 %s 失败: %w", latest, err)
		}
		fmt.Printf("使用数据文件: %s\n", filepath.Base(latest))
		allStatements = append(allStatements, resp.FlexStatements...)
	}

	if len(allStatements) == 0 {
		return nil, fmt.Errorf("无可用数据，请先执行 ibkr fetch")
	}
	return allStatements, nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func reportCmd() *cobra.Command {
	var outputFile string
	cmd := &cobra.Command{
		Use:   "report",
		Short: "生成 Markdown 格式的综合报告",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig()
			if err != nil {
				return err
			}

			statements, err := loadLatestData(cfg)
			if err != nil {
				return err
			}

			md := analysis.GenerateMarkdownReport(statements, flagFrom, flagTo)

			if outputFile == "" {
				outputFile = filepath.Join(cfg.DataDir, fmt.Sprintf("report_%s.md", time.Now().Format("20060102_150405")))
			}
			if err := os.WriteFile(outputFile, []byte(md), 0644); err != nil {
				return fmt.Errorf("保存报告失败: %w", err)
			}
			fmt.Printf("✓ 报告已生成: %s\n", outputFile)
			return nil
		},
	}
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径（默认保存到 data 目录）")
	return cmd
}

func availableQueries(cfg *Config) string {
	var names []string
	for name := range cfg.Queries {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}
