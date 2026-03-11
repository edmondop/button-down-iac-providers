package main

import (
	"context"
	"log"

	"github.com/edmondop/terraform-provider-buttondown/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

const version = "0.1.0"

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/edmondop/buttondown",
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
