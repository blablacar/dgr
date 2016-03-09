package main

import (
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

const INIT_MANIFEST_CONTENT = `name: aci.example.com/aci-dummy:1
`

var initCmd = &cobra.Command{

	Use:   "init",
	Short: "init files-tree",
	Long:  `init files-tree`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		fields := data.WithField("path", workPath)
		if _, err := os.Stat(workPath); err != nil {
			if err := os.MkdirAll(workPath, 0755); err != nil {
				logs.WithEF(err, fields).Fatal("Cannot create path directory")
			}
		}

		empty, err := common.IsDirEmpty(workPath)
		if err != nil {
			logs.WithEF(err, fields).Fatal("Cannot read path directory")
		}
		if !Args.Force {
			if !empty {
				logs.WithEF(err, fields).Fatal("Path is not empty cannot init")
			}
		}

		if err := ioutil.WriteFile(workPath+PATH_ACI_MANIFEST, []byte(INIT_MANIFEST_CONTENT), 0644); err != nil {
			logs.WithEF(err, fields).Fatal("failed to write aci manifest")
		}

		defer giveBackUserRights(workPath)
		NewAciOrPod(workPath, Args).Init()
	},
}

func init() {
	initCmd.Flags().BoolVarP(&Args.Force, "force", "f", false, "Force init command if path is not empty")
}
