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
	Username string `json:"entity_name"`
	RoleID   int    `json:"role_id"`
}

func (u *UserScope) String() string {
	return fmt.Sprintf(`{"member_user":{"username":"%s"},"role_id":%d}`, u.Username, u.RoleID)
}

// GetRepoMemberList 获取指定项目的用户列表
func GetRepoMemberList(id int, repo *config.Repo) ([]*UserScope, error) {
	ret := make([]*UserScope, 0, 100)
	request := gorequest.New()
	// 获取项目内的所有用户
	url := fmt.Sprintf("%s/%s/%d/%s", repo.Addr(), config.PROJECT, id, config.MEMBERS)

	request.Get(url).Set("Content-Type", "application/json").Set("Cookie", "rl_page_init_referrer=RudderEncrypt%3AU2FsdGVkX19BckbrrKwvpg3b8WWKfc00Rj1mkPd6fJI%3D; rl_page_init_referring_domain=RudderEncrypt%3AU2FsdGVkX1%2FCjMq0Z4tvw2IN%2BsfuCK2aYO56u6UJbXs%3D; rl_anonymous_id=RudderEncrypt%3AU2FsdGVkX1%2BnE7X2p1tqaTqrEMbdAHZUZKOl0WeD73Z6vfjzIhp0JVaxKQDzhs4q7%2Fdg6FG4nppWJiBx03pewA%3D%3D; rl_group_id=RudderEncrypt%3AU2FsdGVkX1%2BecwC0ttbg3X8AwCe%2F%2BNlukjy15RGC2eo%3D; rl_group_trait=RudderEncrypt%3AU2FsdGVkX1%2Bmgwx5RFXR8Y%2B3QNarsJxv4KRrKB00WL4%3D; rl_trait=RudderEncrypt%3AU2FsdGVkX1%2Bz0306pJJW9IG9g0K6if2UibKcMzRjF38%3D; rl_user_id=RudderEncrypt%3AU2FsdGVkX19wARoKaqOOHcec037lJMKWADloKSl6soYJ6vP1dYjp1jYlR4SBtvZE; ph_mqkwGT0JNFqO-zX2t0mW6Tec9yooaVu7xCBlXtHnt5Y_posthog=%7B%22distinct_id%22%3A%22181e5cc772929e-0c84ce536986ee-26021a51-e1000-181e5cc772ae59%22%2C%22%24device_id%22%3A%22181e5cc772929e-0c84ce536986ee-26021a51-e1000-181e5cc772ae59%22%7D; Hm_lvt_d53fdbddfd8c982ae1edc0f9da8ed194=1657724439; Hm_cv_d53fdbddfd8c982ae1edc0f9da8ed194=1*username*yuanyupeng; ajs_user_id=%2237e2e4b049d5631f37dbf0147e0803ce3f3e555e%22; ajs_anonymous_id=%228d3b9016-959f-40ec-994b-b031bacc40ff%22; harbor-lang=zh-cn; _gorilla_csrf=MTY2Mjg1Njg5MXxJa1UxTXpWcVZFczNVRVp1YVhFMmEwaE9WMjR6VXpKUVowcHdiblp1VlU5a2R6aDBTbkowTlRkQlF6ZzlJZ289fCM4aDvcrQhiXGxzTuEzkal5U4uGmxoWp7SkIbj8jOl3; sid=d303452e86050776569cbc745db9d3bc") // SetBasicAuth(repo.UserName, repo.Password)

	response, _, err := request.End()
	if err != nil {
		log.Error("get repo member list error", log.Int("project_id", id))
		return nil, err[0]
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get member list from repe: %d failed, http code error: %d", id, response.StatusCode)
	}

	data, _ := io.ReadAll(response.Body)

	if err := json.Unmarshal(data, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// RegisteredMember 注册用户
func RegisteredMember(userScope *UserScope, id int, repo *config.Repo) error {
	request := gorequest.New()
	// 生产url
	url := fmt.Sprintf("%s/%s/%d/%s", repo.Addr(), config.PROJECT, id, config.MEMBERS)
	request.Post(url).Set("Content-Type", "application/json").SetBasicAuth(repo.UserName, repo.Password).Send(userScope.String())
	// fmt.Println(request.Url, request.RawString)
	response, _, errs := request.End()

	if errs != nil {
		log.Error("registered repo member error", log.Int("project_id", id), log.String("username", userScope.Username))
		// fmt.Println("----------", errs)
		return errs[0]
	}

	if response.StatusCode == http.StatusConflict {
		log.Debug("user already in this project", log.Int("project_id", id), log.String("username", userScope.Username))
		return nil
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(response.Body)
		fmt.Println("xxxxxx", string(b))
		fmt.Println(response.StatusCode, response.Status)
		return fmt.Errorf("registered %s to %d error: %d", userScope.Username, id, response.StatusCode)

	}
	return nil
}
