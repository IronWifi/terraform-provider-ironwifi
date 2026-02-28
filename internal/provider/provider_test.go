package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"ironwifi": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProvider_Schema(t *testing.T) {
	// Verify the provider schema compiles and is valid
	p := New("test")()
	if p == nil {
		t.Fatal("provider factory returned nil")
	}
}
