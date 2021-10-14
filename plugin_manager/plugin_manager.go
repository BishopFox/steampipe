package plugin_manager

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/turbot/steampipe-plugin-sdk/logging"

	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/go-plugin"
	"github.com/turbot/go-kit/helpers"
	sdkpluginshared "github.com/turbot/steampipe-plugin-sdk/grpc/shared"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

// PluginManager is the real implementation of grpc.PluginManager
type PluginManager struct {
	pb.UnimplementedPluginManagerServer

	Plugins map[string]*pb.ReattachConfig

	configDir        string
	mut              sync.Mutex
	connectionConfig map[string]*pb.ConnectionConfig
}

func NewPluginManager(connectionConfig map[string]*pb.ConnectionConfig) *PluginManager {
	return &PluginManager{
		connectionConfig: connectionConfig,
		Plugins:          make(map[string]*pb.ReattachConfig),
	}

}

func (m *PluginManager) Serve() {
	// create a plugin map, using ourselves as the implementation
	pluginMap := map[string]plugin.Plugin{
		pluginshared.PluginName: &pluginshared.PluginManagerPlugin{Impl: m},
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginMap,
		//  enable gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// plugin interface functions

func (m *PluginManager) Get(req *pb.GetRequest) (resp *pb.GetResponse, err error) {
	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	log.Printf("[WARN] ****************** PluginManager %p Get connection '%s', plugins %+v\n", m, req.Connection, m.Plugins)

	// is this plugin already running
	if plugin, ok := m.Plugins[req.Connection]; ok {
		log.Printf("[WARN] PluginManager %p found '%s' in map %v", m, req.Connection, m.Plugins)

		// check the pid exists
		exists, _ := utils.PidExists(int(plugin.Pid))
		if exists {
			// return the reattach config
			return &pb.GetResponse{
				Reattach: plugin,
			}, nil
		}
		//log.Printf("[WARN] PluginManager %p plugin pid %d for connection '%s' found in plugin map but does not exist - removing from map", m, plugin.Pid, req.Connection)
		// so there is an entry in the map but it does not exist - remove from the map
		delete(m.Plugins, req.Connection)
	}

	log.Printf("[WARN] PluginManager %p '%s' NOT found in map %v - starting", m, req.Connection, m.Plugins)
	// so we need to start the plugin
	reattach, err := m.startPlugin(req)
	if err != nil {
		return nil, err
	}

	// store the reattach config in our map
	m.Plugins[req.Connection] = reattach

	log.Printf("[WARN] ****************** PluginManager %p Get complete", m)

	// and return
	return &pb.GetResponse{
		Reattach: reattach,
	}, nil

}

func (m *PluginManager) SetConnectionConfigMap(req *pb.SetConnectionConfigMapRequest) (resp *pb.SetConnectionConfigMapResponse, err error) {
	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()
	m.connectionConfig = req.ConfigMap
	return &pb.SetConnectionConfigMapResponse{}, nil
}

func (m *PluginManager) Shutdown(*pb.ShutdownRequest) (resp *pb.ShutdownResponse, err error) {
	m.mut.Lock()
	defer func() {
		m.mut.Unlock()
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	log.Printf("[WARN] Shutdown")
	return &pb.ShutdownResponse{}, nil

	var errs []error
	for _, p := range m.Plugins {
		log.Printf("[WARN] kill %v", p)
		err = m.killPlugin(p.Pid)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return &pb.ShutdownResponse{}, utils.CombineErrorsWithPrefix(fmt.Sprintf("failed to shutdown %d plugins", len(errs)), errs...)
}

func (m *PluginManager) startPlugin(req *pb.GetRequest) (*pb.ReattachConfig, error) {
	// get connection config
	connectionConfig, ok := m.connectionConfig[req.Connection]
	if !ok {
		return nil, fmt.Errorf("no config loaded for connection %s", req.Connection)
	}

	pluginPath, err := GetPluginPath(connectionConfig.Plugin, connectionConfig.PluginShortName)
	if err != nil {
		return nil, err
	}

	// create the plugin map
	pluginName := connectionConfig.Plugin
	pluginMap := map[string]plugin.Plugin{
		pluginName: &sdkpluginshared.WrapperPlugin{},
	}
	loggOpts := &hclog.LoggerOptions{Name: "plugin"}
	if req.DisableLogger {
		loggOpts.Exclude = func(hclog.Level, string, ...interface{}) bool { return true }
	}
	logger := logging.NewLogger(loggOpts)

	cmd := exec.Command(pluginPath)
	// pass env to command
	cmd.Env = os.Environ()
	// set attributes on the command to ensure the process is not shutdown when its parent terminates
	cmd.SysProcAttr = &syscall.SysProcAttr{
		//Setpgid:    true,
		Foreground: false,
	}
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  sdkpluginshared.Handshake,
		Plugins:          pluginMap,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})

	if _, err := client.Start(); err != nil {
		return nil, err
	}
	reattach := client.ReattachConfig()
	return pb.NewReattachConfig(reattach), nil

	/* hub did this

	// loop as we may need to retry if the plugin exists in the map but has actually exited
	const maxAttempts = 3
	for attempt := 1; attempt < maxAttempts; attempt++ {
		// ask connection map to get or create this connection
		c, err := h.connections.get(pluginFQN, connectionName)
		if err != nil {
			return nil, err
		}

		// make sure that the plugin is running
		// (i.e. it has not crashed)
		if !c.Plugin.Client.Exited() {
			// it is running, return it
			return c, nil
		}

		// remove connection from the connection map and kill the GRPC client
		h.connections.removeAndKill(pluginFQN, connectionName)
	}
	*/
}

func (m *PluginManager) killPlugin(pid int64) error {
	process, err := utils.FindProcess(int(pid))
	if err != nil {
		log.Printf("[WARN] error finding process %d", pid)
		return err
	}
	if process == nil {
		return nil
	}
	return process.Kill()
}