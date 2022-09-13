package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/yuanyp8/log"
	"github.com/yuanyp8/synker4harbor/config"
	"github.com/yuanyp8/synker4harbor/core"
)

var configFile string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start sync members",
	Long:  "start sync members",
	RunE:  start,
}

func start(cmd *cobra.Command, args []string) error {
	// load config pathfile
	defer log.Sync()
	if err := config.C().LoadConf(configFile); err != nil {
		log.Error("load file error")
		return err
	}
	log.Debug("load config success")

	// 获取 src repo list
	srcData := core.NewRepoList()
	if err := srcData.GetData(config.C().SourceRepo); err != nil {
		log.Error(fmt.Sprintln(err), log.String("repo", "source repo"))
		return err
	}

	log.Debug("finish get src repo items", log.Int("count", len(srcData.List)))

	// 将数据转成字典
	srcDict := srcData.GetMap()

	// 获取新的Repo
	destData := core.NewRepoList()
	if err := destData.GetData(config.C().DestinationRepo); err != nil {
		log.Error(fmt.Sprintln(err), log.String("repo", "destination repo"))
		return err
	}
	log.Debug("finish get dest repo items", log.Int("count", len(destData.List)))

	// 遍历新repo
	for _, repo := range destData.List {
		// 根据名称获取老的权限
		// 过滤到新环境的测试项目
		var srcid int
		v, exist := srcDict[repo.Name]
		if exist {
			srcid = v
		} else {
			continue
		}

		ret, err := core.GetRepoMemberList(srcid, config.C().SourceRepo)
		if err != nil {
			log.Error("get project member failed", log.Int("project_id", srcid))
			fmt.Println(err)
			// 这里有错误
			continue
		}
		// 如果项目成员为空则跳过该项目
		if len(ret) == 0 {
			continue
		}
		// 获取项目成员的username以及role_id， 关联新老产品的则为project name
		for _, v := range ret {
			// 注册
			if err := core.RegisteredMember(v, repo.ID, config.C().DestinationRepo); err != nil {
				log.Error("registered user to project failed", log.String("project_name", repo.Name), log.Int("project_id", repo.ID), log.String("username", v.Username), log.Int("role_id", v.RoleID))
			}
			log.Debug("registered user to project successful", log.String("project_name", repo.Name), log.Int("project_id", repo.ID), log.String("username", v.Username), log.Int("role_id", v.RoleID))
		}
	}
	fmt.Println("----done---")
	return nil
}

func init() {
	startCmd.PersistentFlags().StringVarP(&configFile, "config", "f", "/etc/synker.yaml", "config file pathname for synker")
}
