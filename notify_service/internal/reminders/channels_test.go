package reminders

import (
	"context"
	"testing"
)

type testDeliveryChannel struct {
	code string
}

func (c testDeliveryChannel) Code() string { return c.code }

func (c testDeliveryChannel) Deliver(_ context.Context, _ DeliveryRequest) (DeliveryResult, error) {
	return DeliveryResult{}, nil
}

func TestChannelRegistryAddsAdapterWithoutSchedulerChanges(t *testing.T) {
	t.Parallel()

	registry, err := NewChannelRegistry(InAppChannel{}, testDeliveryChannel{code: "test_channel"})
	if err != nil {
		t.Fatalf("NewChannelRegistry(): %v", err)
	}
	channels, err := registry.Resolve([]string{"test_channel", ChannelCodeInApp, "test_channel"})
	if err != nil {
		t.Fatalf("Resolve(): %v", err)
	}
	if len(channels) != 2 || channels[0].Code() != ChannelCodeInApp || channels[1].Code() != "test_channel" {
		t.Fatalf("resolved channels = %#v", channels)
	}
}

func TestChannelRegistryRejectsUnknownAndDuplicateAdapters(t *testing.T) {
	t.Parallel()

	if _, err := NewChannelRegistry(testDeliveryChannel{code: "same"}, testDeliveryChannel{code: "same"}); err == nil {
		t.Fatal("duplicate channel registration unexpectedly succeeded")
	}
	registry, err := NewChannelRegistry(InAppChannel{})
	if err != nil {
		t.Fatalf("NewChannelRegistry(): %v", err)
	}
	if _, err := registry.Resolve([]string{"telegram"}); err == nil {
		t.Fatal("unregistered channel unexpectedly resolved")
	}
	if _, err := NewChannelRegistry(testDeliveryChannel{code: "Bad-Code"}); err == nil {
		t.Fatal("invalid channel code unexpectedly registered")
	}
}
