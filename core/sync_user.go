package core

import (
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/yuanyp8/log"
	"github.com/yuanyp8/synker4harbor/config"
	"io"
	"net/http"
)

// get all repositories information
type Repo struct {
	Name string `json:"name"`
	ID   int    `json:"project_id"`
}

type RepoList struct {
	List []*Repo
}

func NewRepoList() *RepoList {
	return &RepoList{List: make([]*Repo, 0, 1000)}
}

func (l *RepoList) GetData(repo *config.Repo) error {
	request := gorequest.New()
	request.Get(fmt.Sprintf("%s/%s", repo.Addr(), config.PROJECT)).Set("Content-Type", "application/json").SetBasicAuth(repo.UserName, repo.Password)
	request.QueryData.Set("page_size", "100")
	// request.Header.Set()

	// 解决分页分表操作
	var i int = 1
	for {
		request.QueryData.Set("page", fmt.Sprintf("%d", i))

		response, _, err := request.End()
		if err != nil {
			log.Error("get repo list error", log.String("url", fmt.Sprintf("%s/%s", repo.Addr(), config.PROJECT)))
			return err[0]
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("repsonse http code error: %d", response.StatusCode)
		}

		data, _ := io.ReadAll(response.Body)

		if len(data) == 3 {
			log.Debug("the pages already read finished")
			break
		}

		onceRepo := make([]*Repo, 0, 1000)
		if err := json.Unmarshal(data, &onceRepo); err != nil {
			return err
		}
		l.List = append(l.List, onceRepo...)
		i++
	}

	return nil
}

func (l *RepoList) GetMap() map[string]int {
	ret := make(map[string]int)
	for _, repo := range l.List {
		ret[repo.Name] = repo.ID
	}
	return ret
}

type UserScope struct {
	Username string `json:"username"`
	RoleID   int    `json:"role_id"`
}

// GetRepoMemberList 获取指定项目的用户列表
func GetRepoMemberList(id int, repo *config.Repo) ([]*UserScope, error) {
	ret := make([]*UserScope, 0, 100)
	request := gorequest.New()
	// 获取项目内的所有用户
	request.Get(fmt.Sprintf("%s/%s/%d/%s", repo.Addr(), config.PROJECT, id, config.MEMBERS)).Set("Content-Type", "application/json").SetBasicAuth(repo.UserName, repo.Password)

	response, _, err := request.End()
	if err != nil {
		log.Error("get repo member list error", log.Int("project_id", id))
		return nil, err[0]
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("repsonse http code error: %d", response.StatusCode)
	}

	data, _ := io.ReadAll(response.Body)

	if err := json.Unmarshal(data, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
