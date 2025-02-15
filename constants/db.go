package constants

import (
	"fmt"

	"github.com/turbot/steampipe/schema"
)

// dbClient constants
const (
	// MaxParallelClientInits is the number of clients to initialize in parallel
	// if we start initializing all clients together, it leads to bad performance on all
	MaxParallelClientInits = 3
)

// DatabaseListenAddresses is an arrays is listen addresses which Steampipe accepts
var DatabaseListenAddresses = []string{"localhost", "127.0.0.1"}

const (
	DatabaseDefaultPort   = 9193
	DatabaseSuperUser     = "root"
	DatabaseUser          = "steampipe"
	DatabaseName          = "steampipe"
	DatabaseUsersRole     = "steampipe_users"
	DefaultMaxConnections = 5
)

// constants for installing db and fdw images
const (
	DatabaseVersion = "12.1.0"
	FdwVersion      = "0.4.0-rc.3"

	// PostgresImageRef is the OCI Image ref for the databse binaries
	// The 12.1.0 image uses the older jar format 12.1.0-v2 is the same version of postgres,
	// just packaged as gzipped tar files (consistent with oras, faster to unzip).  Once everyone is
	// on a newer build, we can delete the old image move the 12.1.0 tag to the new image, and
	// change this back for consistency
	PostgresImageRef = "us-docker.pkg.dev/steampipe/steampipe/db:12.1.0-v2"
	FdwImageRef      = "us-docker.pkg.dev/steampipe/steampipe/fdw:" + FdwVersion
)

// schema names
const (
	// FunctionSchema is the schema container for all steampipe helper functions
	FunctionSchema = "internal"

	// CommandSchema is the schema which is used to send commands to the FDW
	CommandSchema               = "steampipe_command"
	CacheCommandTable           = "cache"
	CacheCommandOperationColumn = "operation"
	CommandCacheOn              = "cache_on"
	CommandCacheOff             = "cache_off"
	CommandCacheClear           = "cache_clear"
)

// Functions :: a list of SQLFunc objects that are installed in the db 'internal' schema startup
var Functions = []schema.SQLFunc{
	{
		Name:     "glob",
		Params:   map[string]string{"input_glob": "text"},
		Returns:  "text",
		Language: "plpgsql",
		Body: `
declare
	output_pattern text;
begin
	output_pattern = replace(input_glob, '*', '%');
	output_pattern = replace(output_pattern, '?', '_');
	return output_pattern;
end;
`,
	},
}

var ReservedConnectionNames = []string{
	"public",
	FunctionSchema,
}

// introspection table names
const (
	IntrospectionTableQuery              = "steampipe_query"
	IntrospectionTableControl            = "steampipe_control"
	IntrospectionTableBenchmark          = "steampipe_benchmark"
	IntrospectionTableMod                = "steampipe_mod"
	IntrospectionTableDashboard          = "steampipe_dashboard"
	IntrospectionTableDashboardContainer = "steampipe_dashboard_container"
	IntrospectionTableDashboardCard      = "steampipe_dashboard_card"
	IntrospectionTableDashboardChart     = "steampipe_dashboard_chart"
	IntrospectionTableDashboardHierarchy = "steampipe_dashboard_hierarchy"
	IntrospectionTableDashboardImage     = "steampipe_dashboard_image"
	IntrospectionTableDashboardInput     = "steampipe_dashboard_input"
	IntrospectionTableDashboardTable     = "steampipe_dashboard_table"
	IntrospectionTableDashboardText      = "steampipe_dashboard_text"
	IntrospectionTableVariable           = "steampipe_variable"
	IntrospectionTableReference          = "steampipe_reference"
)

// Invoker is a pseudoEnum for the command/operation which starts the service
type Invoker string

const (
	// InvokerService is set when invoked by `service start`
	InvokerService Invoker = "service"
	// InvokerQuery is set when invoked by query command
	InvokerQuery = "query"
	// InvokerCheck is set when invoked by check command
	InvokerCheck = "check"
	// InvokerPlugin is set when invoked by a plugin command
	InvokerPlugin = "plugin"
	// InvokerDashboard is set when invoked by dashboard command
	InvokerDashboard = "dashboard"
	// InvokerConnectionWatcher is set when invoked by the connection watcher process
	InvokerConnectionWatcher = "connection-watcher"
)

// IsValid is a validator for Invoker known values
func (i Invoker) IsValid() error {
	switch i {
	case InvokerService, InvokerQuery, InvokerCheck, InvokerPlugin, InvokerDashboard:
		return nil
	}
	return fmt.Errorf("invalid invoker. Can be one of '%v', '%v', '%v', '%v' or '%v' ", InvokerService, InvokerQuery, InvokerPlugin, InvokerCheck, InvokerDashboard)
}
