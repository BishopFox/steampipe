package modconfig

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// Control :: struct representing the control mod resource
type Control struct {
	ShortName string
	FullName  string `cty:"name"`

	Description   *string           `cty:"description" hcl:"description" column_type:"text"`
	Documentation *string           `cty:"documentation" hcl:"documentation" column_type:"text"`
	Labels        *[]string         `cty:"labels" hcl:"labels" column_type:"jsonb"`
	Links         *[]string         `cty:"links" hcl:"links" column_type:"jsonb"`
	ParentName    *ControlGroupName `cty:"parent" hcl:"parent" column_type:"text"`

	SQL      *string `cty:"sql" hcl:"sql" column_type:"text"`
	Severity *string `cty:"severity" hcl:"severity" column_type:"text"`
	Title    *string `cty:"title" hcl:"title" column_type:"text"`

	DeclRange hcl.Range

	parent   ControlTreeItem
	metadata *ResourceMetadata
}

func NewControl(block *hcl.Block) *Control {
	control := &Control{
		ShortName: block.Labels[0],
		FullName:  fmt.Sprintf("control.%s", block.Labels[0]),
		DeclRange: block.DefRange,
	}
	return control
}

func (c *Control) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

func (c *Control) String() string {
	var labels []string
	if c.Labels != nil {
		labels = *c.Labels
	}
	var links []string
	if c.Links != nil {
		links = *c.Links
	}
	return fmt.Sprintf(`
  -----
  Name: %s
  Title: %s
  Description: %s
  SQL: %s
  Parent: %s
  Labels: %v
  Links: %v
`,
		c.FullName,
		types.SafeString(c.Title),
		types.SafeString(c.Description),
		types.SafeString(c.SQL),
		c.GetParentName(),
		labels, links)
}

// AddChild  :: implementation of ControlTreeItem - controls cannot have children so just return error
func (c *Control) AddChild(child ControlTreeItem) error {
	return errors.New("cannot add child to a control")
}

// GetParentName :: implementation of ControlTreeItem
func (c *Control) GetParentName() string {
	if c.ParentName == nil {
		return ""
	}
	return c.ParentName.Name
}

// SetParent :: implementation of ControlTreeItem
func (c *Control) SetParent(parent ControlTreeItem) error {
	c.parent = parent
	return nil
}

// Name :: implementation of ControlTreeItem, HclResource
// return name in format: 'control.<shortName>'
func (c *Control) Name() string {
	return c.FullName
}

// QualifiedName :: name in format: '<modName>.control.<shortName>'
func (c *Control) QualifiedName() string {
	return fmt.Sprintf("%s.%s", c.metadata.ModShortName, c.FullName)
}

// Path :: implementation of ControlTreeItem
func (c *Control) Path() []string {
	path := []string{c.FullName}
	if c.parent != nil {
		path = append(c.parent.Path(), path...)
	}
	return path
}

// GetMetadata :: implementation of HclResource
func (c *Control) GetMetadata() *ResourceMetadata {
	return c.metadata
}

// SetMetadata :: implementation of HclResource
func (c *Control) SetMetadata(metadata *ResourceMetadata) {
	c.metadata = metadata
}