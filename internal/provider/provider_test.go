package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProviderSchema(t *testing.T) {
	p := New("test")()

	var req provider.SchemaRequest
	var resp provider.SchemaResponse

	p.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{"base_url", "username", "password", "insecure"}
	for _, attr := range expectedAttrs {
		if _, ok := attrs[attr]; !ok {
			t.Errorf("expected attribute %q in provider schema", attr)
		}
	}
}

func TestProviderResourcesAndDataSources(t *testing.T) {
	p := New("test")()

	resources := p.Resources(context.Background())
	if len(resources) != 3 {
		t.Errorf("expected 3 resources, got %d", len(resources))
	}

	dataSources := p.DataSources(context.Background())
	if len(dataSources) != 3 {
		t.Errorf("expected 3 data sources, got %d", len(dataSources))
	}
}
