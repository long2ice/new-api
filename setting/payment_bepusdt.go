package setting

// BEPUSDT hosted checkout configuration.
// Gateway is enabled once BepusdtEnabled is true and Url + ApiKey are set.
var (
	BepusdtEnabled    bool
	BepusdtUrl        string
	BepusdtApiKey     string
	BepusdtFiat       string = "USD"
	BepusdtCurrencies string
	BepusdtTradeType  string
	BepusdtUnitPrice  float64 = 1.0
	BepusdtMinTopUp   int     = 1
	BepusdtNotifyUrl  string
	BepusdtReturnUrl  string
)
