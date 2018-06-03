package moved

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/fsnotify/fsnotify"
	"github.com/genzai-io/sliced/common/pid"
	"github.com/genzai-io/sliced/common/raft"
	"github.com/genzai-io/sliced/proto/store"
	"github.com/rcrowley/go-metrics"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ErrClusterAddress = errors.New("cluster address not set")

	// ErrNotFound is returned when an value or idx is not in the database.
	ErrNotFound = errors.New("not found")

	// ErrInvalid is returned when the database file is an invalid format.
	ErrInvalid = errors.New("invalid database")
	// ErrIndexExists is returned when an idx already exists in the database.
	ErrIndexExists = errors.New("idx exists")

	// ErrInvalidOperation is returned when an operation cannot be completed.
	ErrInvalidOperation = errors.New("invalid operation")

	ErrLogNotShrinkable = errors.New("log not shrinkable")
	ErrNotLeader        = errors.New("not leader")
)

var (
	GIT = ""

	// App stuff
	Name       = "moved"
	VersionStr = "0.1.0-1"
	Version    semver.Version

	// Environment
	InstanceID = ""
	Region     = ""

	// Logging stuff
	Console  bool
	Logger   = CLILogger()
	LogLevel int

	// Metrics stuff
	Metrics = metrics.DefaultRegistry

	// Network stuff
	ApiHost    = ":9002"
	ApiPort    = 9002
	ApiAddr    *net.TCPAddr
	EventLoops = 1
	WebHost    = ":9003"

	RaftHost = ":9004"

	Bootstrap bool

	// Raft stuff
	ClusterAddress = raft.ServerAddress(ApiHost)
	ClusterID      = raft.ServerID(ApiHost)

	RaftTimeout = time.Second * 10

	// File system stuff
	UserHomeDir    = ""
	HomeDir        = ""
	StoreDir       = ""
	ClusterDir     = ""
	SchemaDir      = ""
	SliceStorePath = ""
	DataDir        = ""
	PIDName        = ""

	PathMode os.FileMode = 0755

	PID     *single.Single
	PIDLock *single.LockResult
)

var first = true

func init() {
	// Find Home directory
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	UserHomeDir = usr.HomeDir
	HomeDir = filepath.Join(usr.HomeDir, ".moved", "0")
	DataDir = filepath.Join(HomeDir, "data")
	StoreDir = filepath.Join(HomeDir, "store")
	PIDName = "0"

	// Parse App version
	Version, _ = semver.Make(VersionStr)
}

func ensureDir(path string) {
	if path == "" {
		return
	}
	if err := os.MkdirAll(path, PathMode); err != nil {
		if err != os.ErrExist {
			Logger.Panic().AnErr("err", err).Msgf("os.MkdirAll(\"%s\", %d) error", path, PathMode)
		}
	}
}

func layoutDisk() {
	if StoreDir != ":memory:" {
		ensureDir(StoreDir)
		ClusterDir = filepath.Join(StoreDir, "cluster")
		SchemaDir = filepath.Join(SchemaDir, "schema")
		ensureDir(ClusterDir)
		ensureDir(SchemaDir)
	} else {
		ClusterDir = ":memory:"
		SchemaDir = ":memory:"
	}

	if DataDir != ":memory:" {
		ensureDir(DataDir)
	}
}

func BindCLI(cmd *cobra.Command) {
	// Override config file with CLI flags
	viper.BindPFlag("data.path", cmd.Flags().Lookup("data.path"))
	viper.BindPFlag("store.path", cmd.Flags().Lookup("store.path"))
	viper.BindPFlag("pid", cmd.Flags().Lookup("pid"))
	viper.BindPFlag("web.host", cmd.Flags().Lookup("web.host"))
	viper.BindPFlag("api.host", cmd.Flags().Lookup("api.host"))
	viper.BindPFlag("api.loops", cmd.Flags().Lookup("loops"))
	viper.BindPFlag("bootstrap", cmd.Flags().Lookup("bootstrap"))
	viper.BindPFlag("raft.host", cmd.Flags().Lookup("raft.host"))
}

func CLILogger() zerolog.Logger {
	l := zerolog.New(zerolog.ConsoleWriter{
		Out:     os.Stdout,
		NoColor: false,
	})
	l = l.With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	return l
}

func DaemonLogger(dev bool) zerolog.Logger {
	if dev {
		return CLILogger()
	}

	l := zerolog.New(os.Stdout)
	//l := zerolog.New(zerolog.ConsoleWriter{
	//	Out:     os.Stdout,
	//	NoColor: false,
	//})
	l = l.With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	return l
}

func Configure() error {
	// Change logger to Daemon Logger
	Logger = DaemonLogger(!Console)

	zerolog.SetGlobalLevel(zerolog.Level(LogLevel))
	zerolog.TimeFieldFormat = ""

	return readConfig()
}

func readConfig() error {
	// Setup config defaults
	viper.SetDefault("data.path", DataDir)
	viper.SetDefault("web.host", WebHost)
	viper.SetDefault("api.host", ApiHost)
	viper.SetDefault("api.loops", EventLoops)
	viper.SetDefault("store.path", StoreDir)
	viper.SetDefault("pid", PIDName)
	//viper.SetDefault("raft.host", RaftHost)

	// Setup config file name and directories to search for it
	viper.SetConfigName("config")
	viper.AddConfigPath(".") // Look locally first
	//viper.AddConfigPath("$HOME/.moved") // Then look in user home directory
	//viper.AddConfigPath("/data/config") // Then look in system directory

	// Read the config
	err := viper.ReadInConfig()
	if err != nil {
		switch err.(type) {
		case *viper.ConfigFileNotFoundError, viper.ConfigFileNotFoundError:
			Logger.Warn().Msg("config file was not found. using defaults")
			//Logger.Warn().Err(err)
		}
	} else {
		// Which config file was used?
		Logger.Info().Msgf("config file found: %s", viper.ConfigFileUsed())
	}

	// Apply the config
	if err := applyConfig(); err != nil {
		return err
	}

	// Watch config file for changes
	//viper.OnConfigChange(configChanged)
	//viper.WatchConfig()
	return nil
}

func configChanged(in fsnotify.Event) {
	Logger.Warn().Msgf("config file changed: %s", in.Name)
	applyConfig()
}

//func Lock() *single.LockResult {
//	if PID != nil {
//		return PIDLock
//	}
//
//	PID = single.New(PIDName)
//	result := PID.Lock()
//
//	PIDLock = &result
//	return PIDLock
//}

func Lock() *single.LockResult {
	return &single.LockResult{
		Success: true,
		Err:     nil,
		Pid:     0,
		Port:    0,
	}
}

func Unlock() error {
	return nil
}

//func Unlock() error {
//	if PID == nil {
//		return nil
//	} else {
//		return PID.Unlock()
//	}
//}

func applyConfig() error {
	path := viper.GetString("data.path")
	webHost := viper.GetString("web.host")
	loops := viper.GetInt("api.loops")
	apiHost := viper.GetString("api.host")
	storePath := viper.GetString("store.path")
	PIDName = viper.GetString("pid")
	Bootstrap = viper.GetBool("bootstrap")
	//RaftHost = viper.GetString("raft.host")

	if path == "" {
		path = DataDir
	}
	// Need at least 1 event loop
	if loops <= 0 {
		loops = 1
	}
	// Provide a sane upper event loop limit
	if loops > runtime.NumCPU()*3 {
		loops = runtime.NumCPU() * 3
	}

	if first {
		first = false
		DataDir = path
		EventLoops = loops
		WebHost = webHost
		ApiHost = apiHost
		StoreDir = storePath

		layoutDisk()

		addresses, err := findLocalIPs()
		if err != nil {
			return err
		}

		apiAddr, err := net.ResolveTCPAddr("tcp", ApiHost)
		if err != nil {
			return err
		}
		ApiAddr = apiAddr

		// Was a local interface IP discovered?
		if len(addresses) > 0 {
			// Use the first IP as the Raft address
			ClusterAddress = raft.ServerAddress(fmt.Sprintf("%s:%d", addresses[0].To4(), apiAddr.Port))
			ClusterID = raft.ServerID(ClusterAddress)
			ApiPort = apiAddr.Port

			Logger.Info().Msgf("raft cluster address: %s", ClusterAddress)
			Logger.Info().Msgf("raft cluster id: %s", ClusterID)
		} else {
			// A Raft address could not be determined
			return ErrClusterAddress
		}
	} else {
		// See what changed
		if path != DataDir {
			Logger.Warn().Msgf("path changed from %s to %s but cannot apply this change until a restart", DataDir, path)
		}

		if loops != EventLoops {
			Logger.Warn().Msgf("loops changed from %d to %d but cannot apply this change until a restart", EventLoops, loops)
		}
	}

	return nil
}

func GetDrivesList() []*store.Drive {
	m := GetDrives()
	drives := make([]*store.Drive, len(m))
	for _, d := range m {
		drives = append(drives, d)
	}
	return drives
}

// Retrieves the drives from config
func GetDrives() (drives map[string]*store.Drive) {
	driveMap := viper.GetStringMap("drives")
	drives = make(map[string]*store.Drive)

	var working *store.Drive

loop:
	for k, v := range driveMap {
		drive := &store.Drive{
			Mount: k,
		}

		k = strings.ToLower(k)
		if strings.Contains(k, "nvm") {
			drive.Kind = store.Drive_NVME
		} else if strings.Contains(k, "ssd") {
			drive.Kind = store.Drive_SSD
		} else if strings.Contains(k, "hhd") {
			drive.Kind = store.Drive_HDD
		}

		if k == "home" || k == "/" {
			if working != nil {
				Logger.Warn().Msg("working drive specified more than once")
				continue loop
			}

			working = drive
			drive.Working = true
		}

		switch m := v.(type) {
		case map[string]interface{}:
			for k2, v := range m {
				switch strings.ToLower(k2) {
				case "type", "kind":
					val := fmt.Sprintf("%s", v)
					switch strings.ToLower(val) {
					case "hdd", "hard", "harddrive", "hard-drive":
						drive.Kind = store.Drive_HDD

					case "ssd", "solid-state", "solid":
						drive.Kind = store.Drive_SSD

					case "nvm", "nvme", "pci", "pcie":
						drive.Kind = store.Drive_NVME

					default:
						Logger.Warn().Msgf("invalid drive kind '%s' for drive mounted at: '%s' defaulting to 'hdd'", val, drive.Mount)
						drive.Kind = store.Drive_HDD
					}
				}
				Logger.Debug().Msgf("property: %s -> %s", k2, v)
			}
		default:
			Logger.Debug().Msgf("drive named: %s has an invalid value %v", k, m)
		}

		drives[drive.Mount] = drive
	}

	if working == nil {
		working = &store.Drive{
			Mount:   DataDir,
			Kind:    store.Drive_SSD, // Default to SSD
			Working: true,
		}
		drives[working.Mount] = working
	}

	return
}

func findLocalIPs() ([]net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	addresses := make([]net.IP, 0)
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		// handle err
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.IsLoopback() || v.IP.IsMulticast() {
					continue
				}

				if v.IP.To4() == nil {
					continue
				}

				addresses = append(addresses, v.IP)
			case *net.IPAddr:
				if v.IP.IsLoopback() || v.IP.IsMulticast() {
					continue
				}

				if v.IP.To4() == nil {
					continue
				}

				addresses = append(addresses, v.IP)
			}

		}
	}
	return addresses, nil
}
