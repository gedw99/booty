package components

import (
	"embed"
	"go.amplifyedge.org/booty-v2/internal/store"
	"go.amplifyedge.org/booty-v2/internal/update"
	"io/ioutil"
	"os"
	"path/filepath"

	ks "github.com/kardianos/service"

	"go.amplifyedge.org/booty-v2/dep"
	"go.amplifyedge.org/booty-v2/internal/downloader"
	"go.amplifyedge.org/booty-v2/internal/osutil"
	"go.amplifyedge.org/booty-v2/internal/service"
)

//go:embed files/prometheus.yml
var prometheusCfgSample embed.FS

const (
	vicMetUrlBase = "github.com/VictoriaMetrics/VictoriaMetrics"
	vicmetUrlFmt  = "https://" + vicMetUrlBase
)

type VicMet struct {
	version       update.Version
	db            store.Storer
	svcs          []*service.Svc
	vmStoragePath string
	vmConfigPath  string
}

func (v *VicMet) IsService() bool {
	return true
}

const (
	vname = "victoria-metrics"
)

func NewVicMet(db store.Storer) *VicMet {
	vmStoragePath := filepath.Join(osutil.GetEtcDir(), vname, "storage")
	_ = os.MkdirAll(vmStoragePath, 0700)
	vmConfigPath := filepath.Join(osutil.GetEtcDir(), vname, "prometheus.yml")
	return &VicMet{
		db:            db,
		vmStoragePath: vmStoragePath,
		vmConfigPath:  vmConfigPath,
	}
}

func (v *VicMet) SetVersion(ver update.Version) {
	v.version = ver
}

func (v *VicMet) service() ([]*service.Svc, error) {
	vmConfig := &ks.Config{
		Name:        v.Name(),
		DisplayName: v.Name(),
		Description: "fast, cost-effective monitoring solution and time series database",
		Arguments: []string{
			"--promscrape.config=" + v.vmConfigPath,
			"--opentsdbListenAddr=:4242",
			"--httpListenAddr=:8428",
			"--storageDataPath=" + v.vmStoragePath,
		},
		Executable: filepath.Join(osutil.GetBinDir(), v.Name()),
		Option:     map[string]interface{}{},
	}
	vmAgentConfig := &ks.Config{
		Name:        "vmagent",
		DisplayName: "vmagent",
		Description: "victoria metrics agent",
		Arguments: []string{
			"--remoteWrite.url=http://localhost:8428/api/v1/write",
		},
		Executable: filepath.Join(osutil.GetBinDir(), "vmagent"),
		Option:     map[string]interface{}{},
	}
	vmetsvc, err := service.NewService(vmConfig)
	if err != nil {
		return nil, err
	}
	vmagentSvc, err := service.NewService(vmAgentConfig)
	if err != nil {
		return nil, err
	}
	return []*service.Svc{
		vmetsvc, vmagentSvc,
	}, nil
}

func (v *VicMet) Name() string {
	return vname
}

func (v *VicMet) Version() update.Version {
	if v := commonGetVersion(v, v.db); v != nil {
		return *v
	}
	return v.version
}

func (v *VicMet) Download() error {
	// victoriametrics doesn't support windows yet
	//if osutil.GetOS() == "windows" {
	//	return nil
	//}
	targetDir := getDlPath(v.Name(), v.version.String())
	return downloader.Download(vicMetUrlBase+"?ref=v"+v.version.String(), targetDir)
}

func (v *VicMet) Dependencies() []dep.Component {
	return nil
}

func (v *VicMet) Install() error {
	//if osutil.GetOS() == "windows" {
	//	return nil
	//}
	dlPath := getDlPath(v.Name(), v.version.String())
	var err error
	binDir := osutil.GetBinDir()
	if err = os.Chdir(dlPath); err != nil {
		return err
	}
	recipes := []string{"victoria-metrics", "vmagent", "vmalert", "vmauth", "vmbackup", "vmctl", "vminsert", "vmrestore", "vmselect", "vmstorage"}
	// make pure local binaries
	filesMap := map[string][]interface{}{}
	for _, r := range recipes {
		if err = osutil.Exec("make", "app-local-pure", "APP_NAME="+r); err != nil {
			return err
		}
		filesMap[filepath.Join(dlPath, "bin", r+"-pure")] = []interface{}{
			filepath.Join(binDir, r), 0755,
		}
	}
	ip, err := commonInstall(v, filesMap)
	if err != nil {
		return err
	}

	// install default config
	if exists := osutil.Exists(v.vmConfigPath); !exists {
		promData, err := prometheusCfgSample.ReadFile("files/prometheus.yml")
		if err != nil {
			return err
		}
		if err = ioutil.WriteFile(v.vmConfigPath, promData, 0600); err != nil {
			return err
		}
	}
	// install services
	svcs, err := v.service()
	if err != nil {
		return err
	}
	v.svcs = svcs
	for _, s := range v.svcs {
		_ = s.Install()
	}
	if err = v.db.New(ip); err != nil {
		return err
	}
	return nil
}

func (v *VicMet) Uninstall() error {
	if osutil.GetOS() == "windows" {
		return nil
	}
	var err error
	var pkg *store.InstalledPackage
	pkg, err = v.db.Get(v.Name())
	if err != nil {
		return err
	}
	var filesList []string
	for k := range pkg.FilesMap {
		filesList = append(filesList, k)
	}
	for _, file := range filesList {
		if err = os.RemoveAll(file); err != nil {
			return err
		}
	}

	if v.svcs == nil {
		v.svcs, _ = v.service()
	}
	for _, s := range v.svcs {
		_ = s.Uninstall()
	}
	if err = v.db.Delete(v.Name()); err != nil {
		return err
	}
	_ = os.RemoveAll(getDlPath(v.Name(), v.version.String()))
	return nil
}

func (v *VicMet) Run(args ...string) error {
	if v.svcs == nil {
		svcs, err := v.service()
		if err != nil {
			return err
		}
		v.svcs = svcs
	}
	for _, s := range v.svcs {
		if err := s.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (v *VicMet) Update(version update.Version) error {
	return commonUpdate(v, version)
}

func (v *VicMet) RunStop() error {
	if v.svcs == nil {
		svcs, err := v.service()
		if err != nil {
			return err
		}
		v.svcs = svcs
	}
	for _, s := range v.svcs {
		if err := s.Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (v *VicMet) Backup() error {
	// TODO
	return nil
}

func (v *VicMet) IsDev() bool {
	return true
}

func (v *VicMet) RepoUrl() update.RepositoryURL {
	return vicmetUrlFmt
}
