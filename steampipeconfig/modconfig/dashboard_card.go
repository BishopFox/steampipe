package modconfig

import (
	"fmt"

	"github.com/turbot/steampipe/constants"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/zclconf/go-cty/cty"
)

// DashboardCard is a struct representing a leaf dashboard node
type DashboardCard struct {
	DashboardLeafNodeBase
	ResourceWithMetadataBase
	QueryProviderBase

	FullName        string `cty:"name" json:"-"`
	ShortName       string `json:"-"`
	UnqualifiedName string `json:"-"`

	// these properties are JSON serialised by the parent LeafRun
	Title   *string        `cty:"title" hcl:"title" column:"title,text" json:"-"`
	Width   *int           `cty:"width" hcl:"width" column:"width,text"  json:"-"`
	Type    *string        `cty:"type" hcl:"type" column:"type,text" json:"type,omitempty"`
	Icon    *string        `cty:"icon" hcl:"icon" column:"icon,text" json:"icon,omitempty"`
	Display *string        `cty:"display" hcl:"display" json:"display,omitempty"`
	OnHooks []*DashboardOn `cty:"on" hcl:"on,block" json:"on,omitempty"`

	// QueryProvider
	SQL                   *string     `cty:"sql" hcl:"sql" column:"sql,text" json:"sql"`
	Query                 *Query      `hcl:"query" json:"-"`
	PreparedStatementName string      `column:"prepared_statement_name,text" json:"-"`
	Args                  *QueryArgs  `cty:"args" column:"args,jsonb" json:"args,omitempty"`
	Params                []*ParamDef `cty:"params" column:"params,jsonb" json:"params,omitempty"`

	Base *DashboardCard `hcl:"base" json:"-"`

	DeclRange hcl.Range  `json:"-"`
	Mod       *Mod       `cty:"mod" json:"-"`
	Paths     []NodePath `column:"path,jsonb" json:"-"`

	parents  []ModTreeItem
	metadata *ResourceMetadata
}

func NewDashboardCard(block *hcl.Block, mod *Mod) *DashboardCard {
	shortName := GetAnonymousResourceShortName(block, mod)
	c := &DashboardCard{
		ShortName:       shortName,
		FullName:        fmt.Sprintf("%s.%s.%s", mod.ShortName, block.Type, shortName),
		UnqualifiedName: fmt.Sprintf("%s.%s", block.Type, shortName),
		Mod:             mod,
		DeclRange:       block.DefRange,
	}

	c.SetAnonymous(block)
	return c
}

func (c *DashboardCard) Equals(other *DashboardCard) bool {
	diff := c.Diff(other)
	return !diff.HasChanges()
}

// CtyValue implements HclResource
func (c *DashboardCard) CtyValue() (cty.Value, error) {
	return getCtyValue(c)
}

// Name implements HclResource, ModTreeItem
// return name in format: 'card.<shortName>'
func (c *DashboardCard) Name() string {
	return c.FullName
}

// OnDecoded implements HclResource
func (c *DashboardCard) OnDecoded(*hcl.Block) hcl.Diagnostics {
	c.setBaseProperties()
	return nil
}

func (c *DashboardCard) setBaseProperties() {
	if c.Base == nil {
		return
	}
	if c.Title == nil {
		c.Title = c.Base.Title
	}
	if c.Type == nil {
		c.Type = c.Base.Type
	}
	if c.Icon == nil {
		c.Icon = c.Base.Icon
	}
	if c.Width == nil {
		c.Width = c.Base.Width
	}
	if c.SQL == nil {
		c.SQL = c.Base.SQL
	}
}

// AddReference implements HclResource
func (c *DashboardCard) AddReference(*ResourceReference) {}

// GetMod implements HclResource
func (c *DashboardCard) GetMod() *Mod {
	return c.Mod
}

// GetDeclRange implements HclResource
func (c *DashboardCard) GetDeclRange() *hcl.Range {
	return &c.DeclRange
}

// AddParent implements ModTreeItem
func (c *DashboardCard) AddParent(parent ModTreeItem) error {
	c.parents = append(c.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (c *DashboardCard) GetParents() []ModTreeItem {
	return c.parents
}

// GetChildren implements ModTreeItem
func (c *DashboardCard) GetChildren() []ModTreeItem {
	return nil
}

// GetTitle implements ModTreeItem
func (c *DashboardCard) GetTitle() string {
	return typehelpers.SafeString(c.Title)
}

// GetDescription implements ModTreeItem
func (c *DashboardCard) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (c *DashboardCard) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (c *DashboardCard) GetPaths() []NodePath {
	// lazy load
	if len(c.Paths) == 0 {
		c.SetPaths()
	}

	return c.Paths
}

// SetPaths implements ModTreeItem
func (c *DashboardCard) SetPaths() {
	for _, parent := range c.parents {
		for _, parentPath := range parent.GetPaths() {
			c.Paths = append(c.Paths, append(parentPath, c.Name()))
		}
	}
}

func (c *DashboardCard) Diff(other *DashboardCard) *DashboardTreeItemDiffs {
	res := &DashboardTreeItemDiffs{
		Item: c,
		Name: c.Name(),
	}
	if !utils.SafeStringsEqual(c.FullName, other.FullName) {
		res.AddPropertyDiff("Name")
	}

	if !utils.SafeStringsEqual(c.Title, other.Title) {
		res.AddPropertyDiff("Title")
	}

	if !utils.SafeStringsEqual(c.SQL, other.SQL) {
		res.AddPropertyDiff("SQL")
	}

	if !utils.SafeIntEqual(c.Width, other.Width) {
		res.AddPropertyDiff("Width")
	}

	if !utils.SafeStringsEqual(c.Type, other.Type) {
		res.AddPropertyDiff("Type")
	}

	if !utils.SafeStringsEqual(c.Icon, other.Icon) {
		res.AddPropertyDiff("Icon")
	}

	res.populateChildDiffs(c, other)

	return res
}

// GetWidth implements DashboardLeafNode
func (c *DashboardCard) GetWidth() int {
	if c.Width == nil {
		return 0
	}
	return *c.Width
}

// GetUnqualifiedName implements DashboardLeafNode
func (c *DashboardCard) GetUnqualifiedName() string {
	return c.UnqualifiedName
}

// GetParams implements QueryProvider
func (c *DashboardCard) GetParams() []*ParamDef {
	return c.Params
}

// GetArgs implements QueryProvider
func (c *DashboardCard) GetArgs() *QueryArgs {
	return c.Args
}

// GetSQL implements QueryProvider
func (c *DashboardCard) GetSQL() *string {
	return c.SQL
}

// GetQuery implements QueryProvider
func (c *DashboardCard) GetQuery() *Query {
	return c.Query
}

// SetArgs implements QueryProvider
func (c *DashboardCard) SetArgs(args *QueryArgs) {
	c.Args = args
}

// SetParams implements QueryProvider
func (c *DashboardCard) SetParams(params []*ParamDef) {
	c.Params = params
}

// GetPreparedStatementName implements QueryProvider
func (c *DashboardCard) GetPreparedStatementName() string {
	if c.PreparedStatementName != "" {
		return c.PreparedStatementName
	}
	c.PreparedStatementName = c.buildPreparedStatementName(c.ShortName, c.Mod.NameWithVersion(), constants.PreparedStatementCardSuffix)
	return c.PreparedStatementName
}

// GetPreparedStatementExecuteSQL implements QueryProvider
func (c *DashboardCard) GetPreparedStatementExecuteSQL(args *QueryArgs) (string, error) {
	// defer to base
	return c.getPreparedStatementExecuteSQL(c, args)
}
