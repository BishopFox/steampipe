package workspace

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/steampipeconfig/parse"
	"github.com/turbot/steampipe/steampipeconfig/versionmap"
	"github.com/turbot/steampipe/utils"
)

type Workspace struct {
	Path                string
	ModInstallationPath string
	Mod                 *modconfig.Mod

	Mods      map[string]*modconfig.Mod
	Variables map[string]*modconfig.Variable

	watcher    *utils.FileWatcher
	loadLock   sync.Mutex
	exclusions []string
	// should we load/watch files recursively
	listFlag                filehelpers.ListFlag
	fileWatcherErrorHandler func(context.Context, error)
	watcherError            error
	// event handlers
	dashboardEventHandlers []dashboardevents.DashboardEventHandler
	// callback function to reset display after the file watche displays messages
	onFileWatcherEventMessages func()
	modFileExists              bool
	loadPseudoResources        bool

	// maps of mod resources from this mod and ALL DEPENDENCIES, keyed by long and short names
	resourceMaps *modconfig.WorkspaceResourceMaps
}

// Load creates a Workspace and loads the workspace mod
func Load(ctx context.Context, workspacePath string) (*Workspace, error) {
	utils.LogTime("workspace.Load start")
	defer utils.LogTime("workspace.Load end")

	// create shell workspace
	workspace := &Workspace{
		Path: workspacePath,
	}

	// check whether the workspace contains a modfile
	// this will determine whether we load files recursively, and create pseudo resources for sql files
	workspace.setModfileExists()

	// load the .steampipe ignore file
	if err := workspace.loadExclusions(); err != nil {
		return nil, err
	}

	// load the workspace mod
	if err := workspace.loadWorkspaceMod(ctx); err != nil {
		return nil, err
	}

	// return context error so calling code can handle cancellations
	return workspace, nil
}

// LoadResourceNames builds lists of all workspace resource names
func LoadResourceNames(workspacePath string) (*modconfig.WorkspaceResources, error) {
	utils.LogTime("workspace.LoadResourceNames start")
	defer utils.LogTime("workspace.LoadResourceNames end")

	// create shell workspace
	workspace := &Workspace{
		Path: workspacePath,
	}

	// determine whether to load files recursively or just from the top level folder
	workspace.setModfileExists()

	// load the .steampipe ignore file
	if err := workspace.loadExclusions(); err != nil {
		return nil, err
	}

	return workspace.loadWorkspaceResourceName()
}

func (w *Workspace) SetupWatcher(ctx context.Context, client db_common.Client, errorHandler func(context.Context, error)) error {
	watcherOptions := &utils.WatcherOptions{
		Directories: []string{w.Path},
		Include:     filehelpers.InclusionsFromExtensions(steampipeconfig.GetModFileExtensions()),
		Exclude:     w.exclusions,
		ListFlag:    w.listFlag,
		// we should look into passing the callback function into the underlying watcher
		// we need to analyze the kind of errors that come out from the watcher and
		// decide how to handle them
		// OnError: errCallback,
		OnChange: func(events []fsnotify.Event) {
			w.handleFileWatcherEvent(ctx, client, events)
		},
	}
	watcher, err := utils.NewWatcher(watcherOptions)
	if err != nil {
		return err
	}
	w.watcher = watcher
	// start the watcher
	watcher.Start()

	// set the file watcher error handler, which will get called when there are parsing errors
	// after a file watcher event
	w.fileWatcherErrorHandler = errorHandler
	if w.fileWatcherErrorHandler == nil {
		w.fileWatcherErrorHandler = func(ctx context.Context, err error) {
			fmt.Println()
			utils.ShowErrorWithMessage(ctx, err, "Failed to reload mod from file watcher")
		}
	}

	return nil
}

func (w *Workspace) SetOnFileWatcherEventMessages(f func()) {
	w.onFileWatcherEventMessages = f
}

// access functions
// NOTE: all access functions lock 'loadLock' - this is to avoid conflicts with the file watcher

func (w *Workspace) Close() {
	if w.watcher != nil {
		w.watcher.Close()
	}
}

// check  whether the workspace contains a modfile
// this will determine whether we load files recursively, and create pseudo resources for sql files
func (w *Workspace) setModfileExists() {
	modFilePath := filepaths.ModFilePath(w.Path)
	_, err := os.Stat(modFilePath)
	modFileExists := err == nil

	if modFileExists {
		log.Printf("[TRACE] modfile exists in workspace folder - creating pseudo-resources and loading files recursively ")
		// only load/watch recursively if a mod sp file exists in the workspace folder
		w.listFlag = filehelpers.FilesRecursive
		w.loadPseudoResources = true
	} else {
		log.Printf("[TRACE] no modfile exists in workspace folder - NOT creating pseudoresources and onnly loading resource files from top level folder")
		w.listFlag = filehelpers.Files
		w.loadPseudoResources = false
	}
}

func (w *Workspace) loadWorkspaceMod(ctx context.Context) error {
	// load and evaluate all variables
	inputVariables, err := w.getAllVariables(ctx)
	if err != nil {
		return err
	}

	// build run context which we use to load the workspace
	runCtx, err := w.getRunContext()
	if err != nil {
		return err
	}
	// add variables to runContext
	runCtx.AddVariables(inputVariables)

	// now load the mod
	m, err := steampipeconfig.LoadMod(w.Path, runCtx)
	if err != nil {
		return err
	}

	// now set workspace properties
	w.Mod = m

	// set variables on workspace
	w.Variables = m.Variables
	w.Mods = runCtx.LoadedDependencyMods
	// NOTE: add in the workspace mod to the dependency mods
	w.Mods[w.Mod.Name()] = w.Mod

	// populate the workspace resource map
	w.populateResourceMaps()

	// verify all runtime dependencies can be resolved
	return w.verifyResourceRuntimeDependencies()
}

// build options used to load workspace
// set flags to create pseudo resources and a default mod if needed
func (w *Workspace) getRunContext() (*parse.RunContext, error) {
	parseFlag := parse.CreateDefaultMod
	if w.loadPseudoResources {
		parseFlag |= parse.CreatePseudoResources
	}
	// load the workspace lock
	workspaceLock, err := versionmap.LoadWorkspaceLock(w.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load installation cache from %s: %s", w.Path, err)
	}

	runCtx := parse.NewRunContext(
		workspaceLock,
		w.Path,
		parseFlag,
		&filehelpers.ListOptions{
			// listFlag specifies whether to load files recursively
			Flags:   w.listFlag,
			Exclude: w.exclusions,
			// only load .sp files
			Include: filehelpers.InclusionsFromExtensions([]string{constants.ModDataExtension}),
		})

	return runCtx, nil
}

func (w *Workspace) loadExclusions() error {
	// default to ignoring hidden files and folders
	w.exclusions = []string{
		fmt.Sprintf("%s/**/.*", w.Path),
		fmt.Sprintf("%s/.*", w.Path),
	}

	ignorePath := filepath.Join(w.Path, filepaths.WorkspaceIgnoreFile)
	file, err := os.Open(ignorePath)
	if err != nil {
		// if file does not exist, just return
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) != 0 && !strings.HasPrefix(line, "#") {
			// add exclusion to the workspace path (to ensure relative patterns work)
			absoluteExclusion := filepath.Join(w.Path, line)
			w.exclusions = append(w.exclusions, absoluteExclusion)
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (w *Workspace) loadWorkspaceResourceName() (*modconfig.WorkspaceResources, error) {
	// build options used to load workspace
	runCtx, err := w.getRunContext()
	if err != nil {
		return nil, err
	}

	workspaceResourceNames, err := steampipeconfig.LoadModResourceNames(w.Path, runCtx)
	if err != nil {
		return nil, err
	}

	// TODO load resource names for dependency mods
	//modsPath := file_paths.WorkspaceModPath(w.Path)
	//dependencyResourceNames, err := w.loadModDependencyResourceNames(modsPath)
	//if err != nil {
	//	return nil, err
	//}

	return workspaceResourceNames, nil
}

func (w *Workspace) verifyResourceRuntimeDependencies() error {
	for _, d := range w.resourceMaps.Dashboards {
		if err := d.BuildRuntimeDependencyTree(w); err != nil {
			return err
		}
	}
	return nil
}
