package customtypes

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"
)

type SimpleMoney struct {
	CentAmount   types.Int64  `tfsdk:"cent_amount"`
	CurrencyCode types.String `tfsdk:"currency_code"`
}

func (sm SimpleMoney) ToNative() platform.Money {
	return platform.Money{
		CentAmount:   int(sm.CentAmount.ValueInt64()),
		CurrencyCode: sm.CurrencyCode.ValueString(),
	}
}

func SimpleMoneyFromTypedMoney(tm platform.TypedMoney) SimpleMoney {
	switch tm := tm.(type) {
	case platform.CentPrecisionMoney:
		return SimpleMoney{
			CentAmount:   types.Int64Value(int64(tm.CentAmount)),
			CurrencyCode: types.StringValue(tm.CurrencyCode),
		}
	}
	panic(fmt.Sprintf("Unknown money type: %T", tm))
}
