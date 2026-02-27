package flex

import "encoding/xml"

// SendRequest 响应
type SendRequestResponse struct {
	XMLName       xml.Name `xml:"FlexStatementResponse"`
	Status        string   `xml:"Status"`
	ReferenceCode string   `xml:"ReferenceCode"`
	URL           string   `xml:"Url"`
	ErrorCode     int      `xml:"ErrorCode"`
	ErrorMessage  string   `xml:"ErrorMessage"`
}

// GetStatement 响应（完整的 Flex Query 数据）
type FlexQueryResponse struct {
	XMLName        xml.Name        `xml:"FlexQueryResponse"`
	QueryName      string          `xml:"queryName,attr"`
	Type           string          `xml:"type,attr"`
	FlexStatements []FlexStatement `xml:"FlexStatements>FlexStatement"`
}

type FlexStatement struct {
	AccountID     string `xml:"accountId,attr"`
	FromDate      string `xml:"fromDate,attr"`
	ToDate        string `xml:"toDate,attr"`
	WhenGenerated string `xml:"whenGenerated,attr"`

	Trades           []Trade             `xml:"Trades>Trade"`
	OpenPositions    []OpenPosition      `xml:"OpenPositions>OpenPosition"`
	CashTransactions []CashTransaction   `xml:"CashTransactions>CashTransaction"`
	CashReport       []CashReportCurrency `xml:"CashReport>CashReportCurrency"`
	CorporateActions []CorporateAction   `xml:"CorporateActions>CorporateAction"`
	Transfers        []Transfer          `xml:"Transfers>Transfer"`
}

type Trade struct {
	Symbol          string  `xml:"symbol,attr"`
	Description     string  `xml:"description,attr"`
	AssetCategory   string  `xml:"assetCategory,attr"`
	Currency        string  `xml:"currency,attr"`
	TradeDate       string  `xml:"tradeDate,attr"`
	DateTime        string  `xml:"dateTime,attr"`
	Quantity        float64 `xml:"quantity,attr"`
	TradePrice      float64 `xml:"tradePrice,attr"`
	Proceeds        float64 `xml:"proceeds,attr"`
	Cost            float64 `xml:"cost,attr"`
	RealizedPnL     float64 `xml:"fifoPnlRealized,attr"`
	Commission      float64 `xml:"ibCommission,attr"`
	CommissionCurr  string  `xml:"ibCommissionCurrency,attr"`
	BuySell         string  `xml:"buySell,attr"`
	OpenCloseInd    string  `xml:"openCloseIndicator,attr"`
	TransactionType string  `xml:"transactionType,attr"`
	NetCash         float64 `xml:"netCash,attr"`
	FxRateToBase    float64 `xml:"fxRateToBase,attr"`
	TransactionID   string  `xml:"transactionID,attr"`
	OrderID         string  `xml:"ibOrderID,attr"`
}

type OpenPosition struct {
	Symbol           string  `xml:"symbol,attr"`
	Description      string  `xml:"description,attr"`
	AssetCategory    string  `xml:"assetCategory,attr"`
	Currency         string  `xml:"currency,attr"`
	Position         float64 `xml:"position,attr"`
	MarkPrice        float64 `xml:"markPrice,attr"`
	CostBasis        float64 `xml:"costBasisPrice,attr"`
	CostBasisMoney   float64 `xml:"costBasisMoney,attr"`
	PositionValue    float64 `xml:"positionValue,attr"`
	FifoPnlUnrealized float64 `xml:"fifoPnlUnrealized,attr"`
	FxRateToBase     float64 `xml:"fxRateToBase,attr"`
	ReportDate       string  `xml:"reportDate,attr"`
}

type CashTransaction struct {
	Symbol          string  `xml:"symbol,attr"`
	Description     string  `xml:"description,attr"`
	Currency        string  `xml:"currency,attr"`
	Amount          float64 `xml:"amount,attr"`
	Type            string  `xml:"type,attr"`
	DateTime        string  `xml:"dateTime,attr"`
	TradeDate       string  `xml:"settleDate,attr"`
	FxRateToBase    float64 `xml:"fxRateToBase,attr"`
	TransactionID   string  `xml:"transactionID,attr"`
}

type CorporateAction struct {
	Symbol        string  `xml:"symbol,attr"`
	Description   string  `xml:"description,attr"`
	DateTime      string  `xml:"dateTime,attr"`
	Type          string  `xml:"type,attr"`
	Quantity      float64 `xml:"quantity,attr"`
	Amount        float64 `xml:"amount,attr"`
	Currency      string  `xml:"currency,attr"`
	TransactionID string  `xml:"transactionID,attr"`
}

// Transfer 转账记录（入金/出金）
type Transfer struct {
	AccountID     string  `xml:"accountId,attr"`
	Currency      string  `xml:"currency,attr"`
	FxRateToBase  float64 `xml:"fxRateToBase,attr"`
	Description   string  `xml:"description,attr"`
	Type          string  `xml:"type,attr"`
	Direction     string  `xml:"direction,attr"`
	Amount        float64 `xml:"amount,attr"`
	DateTime      string  `xml:"dateTime,attr"`
	TransactionID string  `xml:"transactionID,attr"`
}

// CashReport 中的货币明细行（区别于 CashTransaction）
type CashReportCurrency struct {
	AccountID       string  `xml:"accountId,attr"`
	Currency        string  `xml:"currency,attr"`
	LevelOfDetail   string  `xml:"levelOfDetail,attr"`
	FromDate        string  `xml:"fromDate,attr"`
	ToDate          string  `xml:"toDate,attr"`
	Commissions     float64 `xml:"commissions,attr"`
	CommissionsMTD  float64 `xml:"commissionsMTD,attr"`
	CommissionsYTD  float64 `xml:"commissionsYTD,attr"`
	Dividends       float64 `xml:"dividends,attr"`
	DividendsMTD    float64 `xml:"dividendsMTD,attr"`
	DividendsYTD    float64 `xml:"dividendsYTD,attr"`
	WithholdingTax  float64 `xml:"withholdingTax,attr"`
	WithholdingMTD  float64 `xml:"withholdingTaxMTD,attr"`   // 注意：xml 中没有 MTD 后缀
	WithholdingYTD  float64 `xml:"withholdingTaxYTD,attr"`   // 也没有 YTD 后缀
	BrokerInterest  float64 `xml:"brokerInterest,attr"`
	BrokerInterestMTD float64 `xml:"brokerInterestMTD,attr"`
	BrokerInterestYTD float64 `xml:"brokerInterestYTD,attr"`
	OtherFees       float64 `xml:"otherFees,attr"`
	OtherFeesYTD    float64 `xml:"otherFeesYTD,attr"`
	StartingCash    float64 `xml:"startingCash,attr"`
	EndingCash      float64 `xml:"endingCash,attr"`
	EndingSettledCash float64 `xml:"endingSettledCash,attr"`
	DepositWithdrawals float64 `xml:"depositWithdrawals,attr"`
	Deposits        float64 `xml:"deposits,attr"`
	DepositsYTD     float64 `xml:"depositsYTD,attr"`
	Withdrawals     float64 `xml:"withdrawals,attr"`
	WithdrawalsYTD  float64 `xml:"withdrawalsYTD,attr"`
	NetTradesSalesYTD float64 `xml:"netTradesSalesYTD,attr"`
	NetTradesPurchasesYTD float64 `xml:"netTradesPurchasesYTD,attr"`
}
