package deliveryloop

import "testing"

func TestDecideDeliveryProtectsSubscriberUpdates(t *testing.T) {
	tests := []struct {
		name  string
		reply string
		want  DeliveryDecision
	}{
		{name: "approved asset publishes", reply: "deliver", want: DeliveryDecision{SubscriberUpdate: "publish", AssetState: "ready"}},
		{name: "uncertain content stays in processing", reply: "hold for review", want: DeliveryDecision{SubscriberUpdate: "hold", AssetState: "processing"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DecideDelivery(tt.reply); got != tt.want {
				t.Fatalf("DecideDelivery(%q) = %+v, want %+v", tt.reply, got, tt.want)
			}
		})
	}
}
