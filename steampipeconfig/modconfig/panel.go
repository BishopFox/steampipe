package modconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
	"github.com/zclconf/go-cty/cty"
)

// Panel is a struct representing the Report resource
type Panel struct {
	FullName  string `cty:"name"`
	ShortName string

	Title  *string `column:"title,text"`
	Type   *string `column:"type,text"`
	Width  *int    `column:"width,text"`
	Height *int    `column:"height,text"`
	Source *string `column:"source,text"`
	SQL    *string `column:"sql,text"`
	Text   *string `column:"text,text"`
	Panels []*Panel

	DeclRange hcl.Range
	Mod       *Mod `cty:"mod"`

	Children []string   `column:"children,jsonb"`
	Paths    []NodePath `column:"path,jsonb"`

	parents         []ModTreeItem
	metadata        *ResourceMetadata
	UnqualifiedName string
}

func NewPanel(block *hcl.Block) *Panel {
	panel := &Panel{
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("panel.%s", block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("panel.%s", block.Labels[0]),
		DeclRange:       block.DefRange,
	}
	return panel
}

// PanelFromFile creates a panel from a markdown file
func PanelFromFile(modPath, filePath string) (MappableResource, []byte, error) {
	p := &Panel{}
	return p.InitialiseFromFile(modPath, filePath)
}

// InitialiseFromFile implements MappableResource
func (p *Panel) InitialiseFromFile(modPath, filePath string) (MappableResource, []byte, error) {
	// only valid for sql files
	if filepath.Ext(filePath) != constants.MarkdownExtension {
		return nil, nil, fmt.Errorf("Panel.InitialiseFromFile must be called with markdown files only - filepath: '%s'", filePath)
	}

	markdownBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	markdown := string(markdownBytes)

	// get a sluggified version of the filename
	name, err := PseudoResourceNameFromPath(modPath, filePath)
	if err != nil {
		return nil, nil, err
	}
	p.ShortName = name
	p.FullName = fmt.Sprintf("panel.%s", name)
	p.Text = &markdown
	p.Source = utils.ToStringPointer("steampipe.panel.markdown")
	return p, markdownBytes, nil
}

// CtyValue implements HclResource
func (p *Panel) CtyValue() (cty.Value, error) {
	return getCtyValue(p)
}

// Name implements HclResource
// return name in format: 'panel.<shortName>'
func (p *Panel) Name() string {
	return p.FullName
}

// OnDecoded implements HclResource
func (p *Panel) OnDecoded(*hcl.Block) hcl.Diagnostics {
	p.setChildNames()
	return nil
}

func (p *Panel) setChildNames() {
	numChildren := len(p.Panels)
	if numChildren == 0 {
		return
	}
	// set children names
	p.Children = make([]string, numChildren)

	for i, p := range p.Panels {
		p.Children[i] = p.Name()
	}
}

// AddReference implements HclResource
func (p *Panel) AddReference(*ResourceReference) {}

// SetMod implements HclResource
func (p *Panel) SetMod(mod *Mod) {
	p.Mod = mod
	p.UnqualifiedName = p.FullName
	p.FullName = fmt.Sprintf("%s.%s", mod.ShortName, p.FullName)
}

// GetMod implements HclResource
func (p *Panel) GetMod() *Mod {
	return p.Mod
}

// GetDeclRange implements HclResource
func (p *Panel) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

// AddChild implements ModTreeItem
func (p *Panel) AddChild(child ModTreeItem) error {
	switch c := child.(type) {
	case *Panel:
		// avoid duplicates
		if !p.containsPanel(c.Name()) {
			p.Panels = append(p.Panels, c)
		}
	case *Report:
		return fmt.Errorf("panels cannot contain reports")
	}
	return nil
}

// AddParent implements ModTreeItem
func (p *Panel) AddParent(parent ModTreeItem) error {
	p.parents = append(p.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (p *Panel) GetParents() []ModTreeItem {
	return p.parents
}

// GetChildren implements ModTreeItem
func (p *Panel) GetChildren() []ModTreeItem {
	children := make([]ModTreeItem, len(p.Panels))
	for i, p := range p.Panels {
		children[i] = p
	}
	return children
}

// GetTitle implements ModTreeItem
func (p *Panel) GetTitle() string {
	return typehelpers.SafeString(p.Title)
}

// GetDescription implements ModTreeItem
func (p *Panel) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (p *Panel) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (p *Panel) GetPaths() []NodePath {
	// lazy load
	if len(p.Paths) == 0 {
		p.SetPaths()
	}

	return p.Paths
}

// SetPaths implements ModTreeItem
func (p *Panel) SetPaths() {
	for _, parent := range p.parents {
		for _, parentPath := range parent.GetPaths() {
			p.Paths = append(p.Paths, append(parentPath, p.Name()))
		}
	}
}

// GetMetadata implements ResourceWithMetadata
func (p *Panel) GetMetadata() *ResourceMetadata {
	return p.metadata
}

// SetMetadata implements ResourceWithMetadata
func (p *Panel) SetMetadata(metadata *ResourceMetadata) {
	p.metadata = metadata
}

func (p *Panel) Diff(new *Panel) *ReportTreeItemDiffs {
	res := &ReportTreeItemDiffs{
		Item: p,
		Name: p.Name(),
	}
	if typehelpers.SafeString(p.Title) != typehelpers.SafeString(new.Title) {
		res.AddPropertyDiff("Title")
	}
	if typehelpers.SafeString(p.Source) != typehelpers.SafeString(new.Source) {
		res.AddPropertyDiff("Source")
	}
	if typehelpers.SafeString(p.SQL) != typehelpers.SafeString(new.SQL) {
		res.AddPropertyDiff("SQL")
	}
	if typehelpers.SafeString(p.Text) != typehelpers.SafeString(new.Text) {
		res.AddPropertyDiff("Text")
	}
	if typehelpers.SafeString(p.Type) != typehelpers.SafeString(new.Type) {
		res.AddPropertyDiff("Type")
	}
	if p.Width == nil || new.Width == nil {
		if !(p.Width == nil && new.Width == nil) {
			res.AddPropertyDiff("Width")
		}
	} else if *p.Width != *new.Width {
		res.AddPropertyDiff("Width")
	}
	if p.Height == nil || new.Height == nil {
		if !(p.Height == nil && new.Height == nil) {
			res.AddPropertyDiff("Height")
		}
	} else if *p.Height != *new.Height {
		res.AddPropertyDiff("Height")
	}

	res.populateChildDiffs(p, new)

	return res
}

func (p *Panel) containsPanel(name string) bool {
	// does this child already exist
	for _, existingPanel := range p.Panels {
		if existingPanel.Name() == name {
			return true
		}
	}
	return false
}
