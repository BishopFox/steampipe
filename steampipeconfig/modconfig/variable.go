package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/tfdiags"
	"github.com/turbot/steampipe/steampipeconfig/modconfig/var_config"
	"github.com/zclconf/go-cty/cty"
)

// Variable is a struct representing a Variable resource
type Variable struct {
	ResourceWithMetadataBase

	ShortName string
	FullName  string `column:"name,text"`

	Description    string    `column:"description,text"`
	Default        cty.Value `column:"default_value,jsonb"`
	Type           cty.Type  `column:"var_type,text"`
	DescriptionSet bool

	// set after value resolution `column:"value,jsonb"`
	Value                      cty.Value `column:"value,jsonb"`
	ValueSourceType            string    `column:"value_source,text"`
	ValueSourceFileName        string    `column:"value_source_file_name,text"`
	ValueSourceStartLineNumber int       `column:"value_source_start_line_number,integer"`
	ValueSourceEndLineNumber   int       `column:"value_source_end_line_number,integer"`
	DeclRange                  hcl.Range
	ParsingMode                var_config.VariableParsingMode
	Mod                        *Mod

	UnqualifiedName string
}

func NewVariable(v *var_config.Variable, mod *Mod) *Variable {
	return &Variable{
		ShortName:       v.Name,
		Description:     v.Description,
		FullName:        fmt.Sprintf("%s.var.%s", mod.ShortName, v.Name),
		UnqualifiedName: fmt.Sprintf("var.%s", v.Name),
		Default:         v.Default,
		Type:            v.Type,
		ParsingMode:     v.ParsingMode,
		Mod:             mod,

		DeclRange: v.DeclRange,
	}
}

func (v *Variable) Equals(other *Variable) bool {
	return v.ShortName == other.ShortName &&
		v.FullName == other.FullName &&
		v.Description == other.Description &&
		v.Default.RawEquals(other.Default) &&
		v.Value.RawEquals(other.Value)
}

// Name implements HclResource, ResourceWithMetadata
func (v *Variable) Name() string {
	return v.FullName
}

// OnDecoded implements HclResource
func (v *Variable) OnDecoded(*hcl.Block) hcl.Diagnostics { return nil }

// AddReference implements HclResource
func (v *Variable) AddReference(*ResourceReference) {}

// GetMod implements HclResource
func (v *Variable) GetMod() *Mod {
	return v.Mod
}

// CtyValue implements HclResource
func (v *Variable) CtyValue() (cty.Value, error) {
	return v.Default, nil
}

// GetDeclRange implements HclResource
func (v *Variable) GetDeclRange() *hcl.Range {
	return &v.DeclRange
}

// Required returns true if this variable is required to be set by the caller,
// or false if there is a default value that will be used when it isn't set.
func (v *Variable) Required() bool {
	return v.Default == cty.NilVal
}

func (v *Variable) SetInputValue(value cty.Value, sourceType string, sourceRange tfdiags.SourceRange) {
	v.Value = value
	v.ValueSourceType = sourceType
	v.ValueSourceFileName = sourceRange.Filename
	v.ValueSourceStartLineNumber = sourceRange.Start.Line
	v.ValueSourceEndLineNumber = sourceRange.End.Line
}
