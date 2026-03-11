package provider

import (
	"os"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func resolveEnvDefault(val types.String, envVar string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	return os.Getenv(envVar)
}
