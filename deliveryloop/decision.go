package deliveryloop

import "strings"

type DeliveryDecision struct {
	SubscriberUpdate string `json:"subscriber_update"`
	AssetState       string `json:"asset_state"`
}

// DecideDelivery keeps unprocessed content out of the subscriber update.
func DecideDelivery(modelReply string) DeliveryDecision {
	if strings.Contains(strings.ToLower(modelReply), "deliver") {
		return DeliveryDecision{SubscriberUpdate: "publish", AssetState: "ready"}
	}
	return DeliveryDecision{SubscriberUpdate: "hold", AssetState: "processing"}
}
