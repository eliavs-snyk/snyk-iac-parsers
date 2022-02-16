package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

// ModuleCall represents a "module" block in a module or file.
type ModuleCall struct {
	Name string

	//SourceAddr      addrs.ModuleSource
	SourceAddrRaw   string
	SourceAddrRange hcl.Range
	SourceSet       bool

	Config hcl.Body

	//Version VersionConstraint

	Count hcl.Expression
	//ForEach hcl.Expression

	//Providers []PassedProviderConfig

	//DependsOn []hcl.Traversal

	DeclRange hcl.Range
}

func decodeModuleBlock(block *hcl.Block, override bool) (ModuleCall, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	mc := ModuleCall{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	schema := moduleBlockSchema
	if override {
		schema = schemaForOverrides(schema)
	}

	content, remain, moreDiags := block.Body.PartialContent(schema)
	diags = append(diags, moreDiags...)
	mc.Config = remain

	//if !hclsyntax.ValidIdentifier(mc.Name) {
	//	diags = append(diags, &hcl.Diagnostic{
	//		Severity: hcl.DiagError,
	//		Summary:  "Invalid module instance name",
	//		Detail:   badIdentifierDetail,
	//		Subject:  &block.LabelRanges[0],
	//	})
	//}

	if attr, exists := content.Attributes["source"]; exists {
		mc.SourceSet = true
		mc.SourceAddrRange = attr.Expr.Range()
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &mc.SourceAddrRaw)
		diags = append(diags, valDiags...)
		//if !valDiags.HasErrors() {
		//	addr, err := addrs.ParseModuleSource(mc.SourceAddrRaw)
		//	mc.SourceAddr = addr
		//	if err != nil {
		//		// NOTE: In practice it's actually very unlikely to end up here,
		//		// because our source address parser can turn just about any string
		//		// into some sort of remote package address, and so for most errors
		//		// we'll detect them only during module installation. There are
		//		// still a _few_ purely-syntax errors we can catch at parsing time,
		//		// though, mostly related to remote package sub-paths and local
		//		// paths.
		//		switch err := err.(type) {
		//		case *getmodules.MaybeRelativePathErr:
		//			diags = append(diags, &hcl.Diagnostic{
		//				Severity: hcl.DiagError,
		//				Summary:  "Invalid module source address",
		//				Detail: fmt.Sprintf(
		//					"Terraform failed to determine your intended installation method for remote module package %q.\n\nIf you intended this as a path relative to the current module, use \"./%s\" instead. The \"./\" prefix indicates that the address is a relative filesystem path.",
		//					err.Addr, err.Addr,
		//				),
		//				Subject: mc.SourceAddrRange.Ptr(),
		//			})
		//		default:
		//			diags = append(diags, &hcl.Diagnostic{
		//				Severity: hcl.DiagError,
		//				Summary:  "Invalid module source address",
		//				Detail:   fmt.Sprintf("Failed to parse module source address: %s.", err),
		//				Subject:  mc.SourceAddrRange.Ptr(),
		//			})
		//		}
		//	}
		//}
		// NOTE: We leave mc.SourceAddr as nil for any situation where the
		// source attribute is invalid, so any code which tries to carefully
		// use the partial result of a failed config decode must be
		// resilient to that.
	}

	//if attr, exists := content.Attributes["version"]; exists {
	//	var versionDiags hcl.Diagnostics
	//	mc.Version, versionDiags = decodeVersionConstraint(attr)
	//	diags = append(diags, versionDiags...)
	//}

	if attr, exists := content.Attributes["count"]; exists {
		mc.Count = attr.Expr
	}
	//
	//if attr, exists := content.Attributes["for_each"]; exists {
	//	if mc.Count != nil {
	//		diags = append(diags, &hcl.Diagnostic{
	//			Severity: hcl.DiagError,
	//			Summary:  `Invalid combination of "count" and "for_each"`,
	//			Detail:   `The "count" and "for_each" meta-arguments are mutually-exclusive, only one should be used to be explicit about the number of resources to be created.`,
	//			Subject:  &attr.NameRange,
	//		})
	//	}
	//
	//	mc.ForEach = attr.Expr
	//}

	//if attr, exists := content.Attributes["depends_on"]; exists {
	//	deps, depsDiags := decodeDependsOn(attr)
	//	diags = append(diags, depsDiags...)
	//	mc.DependsOn = append(mc.DependsOn, deps...)
	//}

	//TODO handle provider refs
	//if attr, exists := content.Attributes["providers"]; exists {
	//	seen := make(map[string]hcl.Range)
	//	pairs, pDiags := hcl.ExprMap(attr.Expr)
	//	diags = append(diags, pDiags...)
	//	for _, pair := range pairs {
	//		key, keyDiags := decodeProviderConfigRef(pair.Key, "providers")
	//		diags = append(diags, keyDiags...)
	//		value, valueDiags := decodeProviderConfigRef(pair.Value, "providers")
	//		diags = append(diags, valueDiags...)
	//		if keyDiags.HasErrors() || valueDiags.HasErrors() {
	//			continue
	//		}
	//
	//		matchKey := key.String()
	//		if prev, exists := seen[matchKey]; exists {
	//			diags = append(diags, &hcl.Diagnostic{
	//				Severity: hcl.DiagError,
	//				Summary:  "Duplicate provider address",
	//				Detail:   fmt.Sprintf("A provider configuration was already passed to %s at %s. Each child provider configuration can be assigned only once.", matchKey, prev),
	//				Subject:  pair.Value.Range().Ptr(),
	//			})
	//			continue
	//		}
	//
	//		rng := hcl.RangeBetween(pair.Key.Range(), pair.Value.Range())
	//		seen[matchKey] = rng
	//		mc.Providers = append(mc.Providers, PassedProviderConfig{
	//			InChild:  key,
	//			InParent: value,
	//		})
	//	}
	//}
	//
	//var seenEscapeBlock *hcl.Block
	//for _, block := range content.Blocks {
	//	switch block.Type {
	//	case "_":
	//		if seenEscapeBlock != nil {
	//			diags = append(diags, &hcl.Diagnostic{
	//				Severity: hcl.DiagError,
	//				Summary:  "Duplicate escaping block",
	//				Detail: fmt.Sprintf(
	//					"The special block type \"_\" can be used to force particular arguments to be interpreted as module input variables rather than as meta-arguments, but each module block can have only one such block. The first escaping block was at %s.",
	//					seenEscapeBlock.DefRange,
	//				),
	//				Subject: &block.DefRange,
	//			})
	//			continue
	//		}
	//		seenEscapeBlock = block
	//
	//		// When there's an escaping block its content merges with the
	//		// existing config we extracted earlier, so later decoding
	//		// will see a blend of both.
	//		mc.Config = hcl.MergeBodies([]hcl.Body{mc.Config, block.Body})
	//
	//	default:
	//		// All of the other block types in our schema are reserved.
	//		diags = append(diags, &hcl.Diagnostic{
	//			Severity: hcl.DiagError,
	//			Summary:  "Reserved block type name in module block",
	//			Detail:   fmt.Sprintf("The block type name %q is reserved for use by Terraform in a future version.", block.Type),
	//			Subject:  &block.TypeRange,
	//		})
	//	}
	//}

	//TODO handle all the variables, any attribute other
	// than the well-known attributes that are described in the schema can be processed as variables

	return mc, diags
}

var moduleBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "source",
			Required: true,
		},
		{
			Name: "version",
		},
		{
			Name: "count",
		},
		{
			Name: "for_each",
		},
		{
			Name: "depends_on",
		},
		{
			Name: "providers",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "_"}, // meta-argument escaping block

		// These are all reserved for future use.
		{Type: "lifecycle"},
		{Type: "locals"},
		{Type: "provider", LabelNames: []string{"type"}},
	},
}
