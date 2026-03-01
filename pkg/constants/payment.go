package constants

type PayMethod string

const (
	PayMethodApplePay     PayMethod = "apple_pay"
	PayMethodCard         PayMethod = "card"
	PayMethodGooglePay    PayMethod = "google_pay"
	PayMethodIdeal        PayMethod = "ideal"
	PayMethodPayPalWallet PayMethod = "paypal_wallet"
)

type Provider string

const (
	ProviderAdyen      Provider = "adyen"
	ProviderAsiabill   Provider = "asiabill"
	ProviderPayPal     Provider = "paypal"
	ProviderPayflowPro Provider = "payflow_pro"
	ProviderStripe     Provider = "stripe"
)

type RuleType string

const (
	// routing rules
	RuleTypeAdvancedRouting RuleType = "advanced_routing"
	RuleTypeDefaultRouting  RuleType = "default_routing"

	// other rules
	RuleTypeFraudCheck      RuleType = "fraud_check"
	RuleTypePriceAdjustment RuleType = "price_adjustment"
	RuleTypeThreeDsOption   RuleType = "three_ds_option"
)
