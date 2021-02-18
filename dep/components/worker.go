package components

import (
	"go.amplifyedge.org/booty-v2/dep"
	"go.amplifyedge.org/booty-v2/internal/fileutil"
	"go.amplifyedge.org/booty-v2/internal/osutil"
	"go.amplifyedge.org/booty-v2/internal/store"
	"os"
	"path/filepath"
)

type workerType struct {
	jobType      string
	targetDir    string
	dependencies []dep.Component
	errChan      chan error
}

func newWorkerType(jobType string, targetDir string, deps []dep.Component, errChan chan error) *workerType {
	return &workerType{
		jobType:      jobType,
		targetDir:    targetDir,
		dependencies: deps,
		errChan:      errChan,
	}
}

func (w *workerType) do(i int) {
	switch w.jobType {
	case "download":
		w.errChan <- w.dependencies[i].Download()
	case "install":
		w.errChan <- w.dependencies[i].Install()
	case "uninstall":
		w.errChan <- w.dependencies[i].Uninstall()
	}
}

func getDlPath(name, version string) string {
	return filepath.Join(osutil.GetDownloadDir(), name+"-"+version)
}

func commonInstall(c dep.Component, filesMap map[string][]interface{}) (*store.InstalledPackage, error) {
	ip := store.InstalledPackage{
		Name:     c.Name(),
		Version:  c.Version().String(),
		FilesMap: map[string]string{},
	}
	for k, v := range filesMap {
		sum, err := fileutil.Copy(k, v[0].(string))
		if err != nil {
			return nil, err
		}
		installedName := v[0].(string)
		installedMode := v[1].(int)
		if err = os.Chmod(installedName, os.FileMode(installedMode)); err != nil {
			return nil, err
		}
		ip.FilesMap[installedName] = sum
	}
	return &ip, nil
}
